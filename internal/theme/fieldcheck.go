package theme

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"text/template/parse"
)

// maxSuggestionDistance caps how far a field name may be from a real one before
// it stops being offered as a correction. Two edits catches the transpositions
// and single-letter slips a suggestion is worth making for; past that the guess
// says more about the algorithm than about what the author meant.
const maxSuggestionDistance = 2

// dotRoot is the name a template calls the value passed to Execute.
const dotRoot = "$"

// checkFields reports every field a theme's templates read that the template
// contract does not have.
//
// It is deliberately one-sided. A field that resolves is silence, a field that
// certainly does not resolve is a finding, and everything in between is
// silence too: the moment the walk cannot say what the dot is, it stops
// checking rather than guessing. A genuine unknown field already fails the
// build, so this command's value is finding one early and finding all of them,
// and a false positive would trade that for a theme author learning to ignore
// it.
func checkFields(files []*themeFile) []Finding {
	c := &checker{defines: map[string]*parse.Tree{}}
	var entered []*parse.Tree
	for _, file := range files {
		for name, tree := range file.trees {
			if name == file.path {
				// A layout's own body is what cress renders a page through. A
				// partial file's body is registered but never executed, so only the
				// {{ define }} blocks inside it can be entered.
				if file.layout {
					entered = append(entered, tree)
				}
				continue
			}
			c.defines[name] = tree
		}
	}

	for tree, dot := range c.resolveDots(entered) {
		c.walkTree(tree, dot)
	}
	return c.findings
}

// checker walks a theme's parse trees against the template contract.
type checker struct {
	// defines maps a {{ define }} name to its tree, which is how a
	// {{ template }} call is followed to the fragment it enters.
	defines map[string]*parse.Tree
	// calls accumulates what type each {{ template }} call passes its callee.
	// It is filled while dots are still being worked out and read once they
	// settle, which is why collecting and reporting are separate passes.
	calls map[string][]reflect.Type
	// reporting is false while dots are being resolved, so the same walk can
	// gather call sites without accusing a template of anything on the strength
	// of a dot that is still provisional.
	reporting bool
	findings  []Finding
}

// resolveDots works out what the dot is inside every template that can be
// entered, returning only those it is sure about.
//
// Layouts are the certain ones: cress executes them with PageData, so their dot
// is not in question. A fragment's dot is whatever its callers passed it, which
// is usually PageData and legitimately is not:
//
//	{{ range .Nav.Main }}{{ template "nav-link" . }}{{ end }}
//
// enters that fragment with a NavLink. So call sites decide, and a fragment
// reached with two different types, or from a caller whose own dot is unknown,
// is left out rather than checked against a guess.
//
// The rounds are what let a fragment calling a fragment settle: each round is
// derived from the previous one alone, so a callee found late can still correct
// a type an earlier round proposed. It stops when a round changes nothing, and
// the cap is there for a theme whose fragments call each other in a cycle.
func (c *checker) resolveDots(entered []*parse.Tree) map[*parse.Tree]reflect.Type {
	pageData := reflect.TypeOf(PageData{})
	seed := func() map[*parse.Tree]reflect.Type {
		dots := make(map[*parse.Tree]reflect.Type, len(entered))
		for _, tree := range entered {
			dots[tree] = pageData
		}
		return dots
	}

	dots := seed()
	for range len(c.defines) + 1 {
		c.calls = map[string][]reflect.Type{}
		for tree, dot := range dots {
			c.walkTree(tree, dot)
		}

		next := seed()
		for name, passed := range c.calls {
			tree, ok := c.defines[name]
			if !ok {
				continue
			}
			if agreed, ok := agree(passed); ok {
				next[tree] = agreed
			}
		}
		if sameDots(dots, next) {
			break
		}
		dots = next
	}

	// Call sites have said everything they are going to. Dropping the map turns
	// recording into a no-op for the reporting pass that follows.
	c.calls = nil
	c.reporting = true
	return dots
}

// agree reduces what a fragment's call sites passed it to the one type they all
// agree on. A single unknown is enough to give up: a fragment entered with a
// value nobody can name is a fragment nothing can be said about.
func agree(passed []reflect.Type) (reflect.Type, bool) {
	if len(passed) == 0 {
		return nil, false
	}
	for _, typ := range passed {
		if typ == nil || typ != passed[0] {
			return nil, false
		}
	}
	return passed[0], true
}

func sameDots(a, b map[*parse.Tree]reflect.Type) bool {
	if len(a) != len(b) {
		return false
	}
	for tree, typ := range a {
		if b[tree] != typ {
			return false
		}
	}
	return true
}

// scope is what the names in a template body resolve to at one point in it: the
// dot, the root value a template calls "$", and the variables in scope. A nil
// type means the value is real but unnamed, which is the signal to stop
// checking rather than to report.
type scope struct {
	tree *parse.Tree
	dot  reflect.Type
	root reflect.Type
	vars map[string]reflect.Type
}

// child copies a scope for a block that may declare its own variables, so a
// {{ range }} body cannot leak one to the text after {{ end }}.
func (s *scope) child(dot reflect.Type) *scope {
	vars := make(map[string]reflect.Type, len(s.vars))
	for name, typ := range s.vars {
		vars[name] = typ
	}
	return &scope{tree: s.tree, dot: dot, root: s.root, vars: vars}
}

func (c *checker) walkTree(tree *parse.Tree, dot reflect.Type) {
	sc := &scope{tree: tree, dot: dot, root: dot, vars: map[string]reflect.Type{}}
	c.walk(tree.Root, sc)
}

// walk visits one node. A ListNode's children share a scope on purpose, since a
// variable declared partway through a body is in scope for the rest of it.
func (c *checker) walk(node parse.Node, sc *scope) {
	switch n := node.(type) {
	case *parse.ListNode:
		if n == nil {
			return
		}
		for _, child := range n.Nodes {
			c.walk(child, sc)
		}
	case *parse.ActionNode:
		// A bare {{ $x := ... }} binds for the rest of the body it sits in, which
		// is why an action's declarations land in the scope it was given rather
		// than in a copy.
		bind(sc, n.Pipe, c.walkPipe(n.Pipe, sc))
	case *parse.IfNode:
		c.walkBranch(&n.BranchNode, sc, branchIf)
	case *parse.WithNode:
		c.walkBranch(&n.BranchNode, sc, branchWith)
	case *parse.RangeNode:
		c.walkBranch(&n.BranchNode, sc, branchRange)
	case *parse.TemplateNode:
		c.walkTemplateCall(n, sc)
	}
}

// branchKind is how a block rebinds the dot for its body.
type branchKind int

const (
	// branchIf leaves the dot alone: {{ if }} tests a value without entering it.
	branchIf branchKind = iota
	// branchWith enters the value, so the body's dot is what the pipe produced.
	branchWith
	// branchRange enters one element at a time.
	branchRange
)

func (c *checker) walkBranch(n *parse.BranchNode, sc *scope, kind branchKind) {
	piped := c.walkPipe(n.Pipe, sc)

	body := sc.child(sc.dot)
	switch kind {
	case branchWith:
		body.dot = piped
		bind(body, n.Pipe, piped)
	case branchRange:
		body.dot = elemType(piped)
		bindRange(body, n.Pipe, piped)
	case branchIf:
		bind(body, n.Pipe, piped)
	}
	c.walk(n.List, body)

	// The else branch never entered the value, so its dot is the one outside.
	if n.ElseList != nil {
		c.walk(n.ElseList, sc.child(sc.dot))
	}
}

// bind assigns every variable a pipe declares the value the pipe produced.
func bind(sc *scope, pipe *parse.PipeNode, value reflect.Type) {
	if pipe == nil {
		return
	}
	for _, decl := range pipe.Decl {
		sc.vars[decl.Ident[0]] = value
	}
}

// bindRange assigns what a {{ range }} declares, which is not the value it
// ranges over: with two variables the first is the index or key and the second
// the element, and with one it is the element.
func bindRange(sc *scope, pipe *parse.PipeNode, over reflect.Type) {
	if pipe == nil {
		return
	}
	for i, decl := range pipe.Decl {
		bound := elemType(over)
		if len(pipe.Decl) == 2 && i == 0 {
			bound = keyType(over)
		}
		sc.vars[decl.Ident[0]] = bound
	}
}

// walkPipe checks every expression in a pipe and reports what it evaluates to,
// which is nil whenever that cannot be named.
func (c *checker) walkPipe(pipe *parse.PipeNode, sc *scope) reflect.Type {
	if pipe == nil {
		return nil
	}
	var result reflect.Type
	for _, cmd := range pipe.Cmds {
		result = c.walkCommand(cmd, sc)
	}
	// Declarations are bound after the pipe runs, and a {{ range }} rebinds them
	// against the element rather than the whole, which is why the branch does its
	// own binding rather than leaving it here.
	return result
}

// walkCommand checks one command and reports its result type. A command with
// more than one argument is a call, and what a call returns is not something
// this walk tracks, so it checks the arguments and names nothing.
func (c *checker) walkCommand(cmd *parse.CommandNode, sc *scope) reflect.Type {
	for _, arg := range cmd.Args {
		c.walkArg(arg, sc)
	}
	if len(cmd.Args) != 1 {
		return nil
	}
	return c.typeOf(cmd.Args[0], sc)
}

// walkArg checks one argument, reporting any field it names that cannot exist.
func (c *checker) walkArg(arg parse.Node, sc *scope) {
	switch n := arg.(type) {
	case *parse.FieldNode:
		c.resolveChain(sc, sc.dot, n.Ident, n, "")
	case *parse.ChainNode:
		c.walkArg(n.Node, sc)
		c.resolveChain(sc, c.typeOf(n.Node, sc), n.Field, n, "")
	case *parse.VariableNode:
		c.walkVariable(n, sc)
	case *parse.PipeNode:
		c.walkPipe(n, sc)
	}
}

func (c *checker) walkVariable(n *parse.VariableNode, sc *scope) {
	name := n.Ident[0]
	base, ok := sc.vars[name]
	if name == dotRoot {
		base, ok = sc.root, true
	}
	if !ok {
		return
	}
	c.resolveChain(sc, base, n.Ident[1:], n, name)
}

// typeOf names what an expression evaluates to, or nil when it cannot.
func (c *checker) typeOf(node parse.Node, sc *scope) reflect.Type {
	switch n := node.(type) {
	case *parse.DotNode:
		return sc.dot
	case *parse.FieldNode:
		return walkType(sc.dot, n.Ident)
	case *parse.ChainNode:
		return walkType(c.typeOf(n.Node, sc), n.Field)
	case *parse.VariableNode:
		if n.Ident[0] == dotRoot {
			return walkType(sc.root, n.Ident[1:])
		}
		return walkType(sc.vars[n.Ident[0]], n.Ident[1:])
	case *parse.PipeNode:
		return c.walkPipe(n, sc)
	}
	return nil
}

// walkType follows a field chain for its type alone, without reporting. It is
// the quiet half of resolveChain, used where a type is wanted rather than a
// verdict.
func walkType(base reflect.Type, chain []string) reflect.Type {
	for _, name := range chain {
		result, verdict := lookupField(base, name)
		if verdict != fieldFound {
			return nil
		}
		base = result
	}
	return base
}

// resolveChain follows a field chain from base, reporting the first step that
// cannot resolve. It stops at the first problem: once one name in a chain is
// wrong, everything after it is a consequence rather than a second finding.
func (c *checker) resolveChain(sc *scope, base reflect.Type, chain []string, node parse.Node, varName string) {
	for i, name := range chain {
		result, verdict := lookupField(base, name)
		switch verdict {
		case fieldFound:
			base = result
		case fieldUnknown:
			return
		case fieldMissing:
			c.report(sc, node, expr(varName, chain[:i]), expr(varName, chain[:i+1]), base, name)
			return
		}
	}
}

// expr rebuilds the text of a field chain as the author wrote it, so a finding
// quotes something they can search their template for. An empty chain is
// whatever the names were read from: the dot, or the variable itself.
func expr(varName string, chain []string) string {
	if len(chain) == 0 {
		if varName == "" {
			return "."
		}
		return varName
	}
	if varName == "" {
		return "." + strings.Join(chain, ".")
	}
	return varName + "." + strings.Join(chain, ".")
}

func (c *checker) report(sc *scope, node parse.Node, base, text string, baseType reflect.Type, name string) {
	if !c.reporting {
		return
	}
	message := missingMessage(base, text, baseType, name)
	c.findings = append(c.findings, Finding{
		File:    sc.tree.ParseName,
		Line:    lineOf(sc.tree, node),
		Message: message,
	})
}

// missingMessage phrases one unresolvable name.
//
// Reading a field off a list gets its own wording because the fix is not a
// spelling correction: .Nav.Main.Title is somebody reaching past a list for the
// entry inside it, and naming a Go type at them would answer a question they
// did not ask.
func missingMessage(base, text string, baseType reflect.Type, name string) string {
	if isList(baseType) {
		return fmt.Sprintf("%s reads a field of a list; range over %s to reach each entry", text, base)
	}
	message := fmt.Sprintf("%s is not a field of %s", text, typeName(baseType))
	if suggestion := nearestField(baseType, name); suggestion != "" {
		message += fmt.Sprintf(" (did you mean %s?)", suggestion)
	}
	return message
}

func isList(t reflect.Type) bool {
	if t == nil {
		return false
	}
	return t.Kind() == reflect.Slice || t.Kind() == reflect.Array
}

// walkTemplateCall records what a {{ template }} call passes, which is how a
// fragment's dot is worked out, and checks the argument itself.
//
// The call is not followed. The fragment has its own tree and is walked once,
// with the dot its callers agreed on, rather than re-checked at every call
// site.
func (c *checker) walkTemplateCall(n *parse.TemplateNode, sc *scope) {
	if n.Pipe == nil {
		// {{ template "x" }} passes no data, so the fragment's dot is nil and
		// nothing in it can be checked.
		c.record(n.Name, nil)
		return
	}
	c.record(n.Name, c.walkPipe(n.Pipe, sc))
}

func (c *checker) record(name string, passed reflect.Type) {
	if c.calls == nil {
		return
	}
	c.calls[name] = append(c.calls[name], passed)
}

// fieldVerdict is what a name turned out to be on a type.
type fieldVerdict int

const (
	// fieldFound means the name resolves; the type may still be unnameable.
	fieldFound fieldVerdict = iota
	// fieldMissing means the base type is one whose members are known, and this
	// is not one of them. Only this verdict becomes a finding.
	fieldMissing
	// fieldUnknown means the base is not a type whose members can be listed, so
	// nothing can be said either way.
	fieldUnknown
)

// lookupField resolves one name against a type the way a template would.
//
// A map is the reason this is three-valued rather than two. PageView.Meta is a
// map[string]any holding a page's front matter, so .Page.Meta.anything is
// correct by design and its type is whatever the author put there. Treating an
// unnameable result as a failure would report the one field cress documents as
// an escape hatch.
func lookupField(base reflect.Type, name string) (reflect.Type, fieldVerdict) {
	if base == nil {
		return nil, fieldUnknown
	}
	if method, ok := methodResult(base, name); ok {
		return method, fieldFound
	}
	for base.Kind() == reflect.Pointer {
		base = base.Elem()
	}

	switch base.Kind() {
	case reflect.Struct:
		field, ok := base.FieldByName(name)
		if !ok || field.PkgPath != "" {
			return nil, fieldMissing
		}
		return nameable(field.Type), fieldFound
	case reflect.Map:
		if base.Key().Kind() != reflect.String {
			return nil, fieldUnknown
		}
		return nameable(base.Elem()), fieldFound
	case reflect.Interface:
		return nil, fieldUnknown
	default:
		// Every remaining kind is one a template cannot take a field of at all,
		// so .Nav.Main.Title (a slice) or .Page.HTML.Text (a string) is as wrong
		// as a misspelling and worth saying so.
		return nil, fieldMissing
	}
}

// methodResult reports the type a method call yields, looking at both the type
// and its pointer since a template calls either.
func methodResult(base reflect.Type, name string) (reflect.Type, bool) {
	method, ok := base.MethodByName(name)
	if !ok && base.Kind() != reflect.Pointer {
		method, ok = reflect.PointerTo(base).MethodByName(name)
	}
	if !ok {
		return nil, false
	}
	if method.Type.NumOut() == 0 {
		return nil, true
	}
	return nameable(method.Type.Out(0)), true
}

// nameable reduces a type to one worth carrying forward, turning the ones whose
// members cannot be listed into the nil that means "stop checking here".
func nameable(t reflect.Type) reflect.Type {
	if t != nil && t.Kind() == reflect.Interface {
		return nil
	}
	return t
}

// elemType is what ranging over a value yields, or nil when that is not
// something this walk can name.
func elemType(t reflect.Type) reflect.Type {
	if t == nil {
		return nil
	}
	switch t.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		return nameable(t.Elem())
	default:
		return nil
	}
}

// keyType is what the first variable of a two-variable {{ range }} is bound to.
func keyType(t reflect.Type) reflect.Type {
	if t == nil {
		return nil
	}
	switch t.Kind() {
	case reflect.Map:
		return nameable(t.Key())
	case reflect.Slice, reflect.Array:
		return reflect.TypeOf(0)
	default:
		return nil
	}
}

// typeName is what a finding calls a type. The package qualifier is dropped
// because a theme author writes templates rather than Go, and "PageView" is the
// name the theme guide gives the thing they are looking at.
func typeName(t reflect.Type) string {
	if t == nil {
		return "the template data"
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if name := t.Name(); name != "" {
		return name
	}
	return t.String()
}

// nearestField offers the real field name closest to what was written, or
// nothing when the closest is too far to be a plausible slip.
func nearestField(base reflect.Type, name string) string {
	candidates := fieldNames(base)
	best, bestDistance := "", maxSuggestionDistance+1
	for _, candidate := range candidates {
		if d := distance(strings.ToLower(candidate), strings.ToLower(name)); d < bestDistance {
			best, bestDistance = candidate, d
		}
	}
	if bestDistance > maxSuggestionDistance {
		return ""
	}
	return best
}

// fieldNames lists what a template could have written instead, sorted so that
// two equally close candidates always yield the same suggestion.
func fieldNames(base reflect.Type) []string {
	if base == nil {
		return nil
	}
	for base.Kind() == reflect.Pointer {
		base = base.Elem()
	}
	if base.Kind() != reflect.Struct {
		return nil
	}
	names := make([]string, 0, base.NumField()+base.NumMethod())
	for i := range base.NumField() {
		if field := base.Field(i); field.PkgPath == "" {
			names = append(names, field.Name)
		}
	}
	for i := range base.NumMethod() {
		names = append(names, base.Method(i).Name)
	}
	sort.Strings(names)
	return names
}

// distance is the Levenshtein edit distance between a and b.
func distance(a, b string) int {
	previous := make([]int, len(b)+1)
	current := make([]int, len(b)+1)
	for j := range previous {
		previous[j] = j
	}
	for i := 1; i <= len(a); i++ {
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			current[j] = min(previous[j]+1, current[j-1]+1, previous[j-1]+cost)
		}
		previous, current = current, previous
	}
	return previous[len(b)]
}

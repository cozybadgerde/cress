package theme

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"
	"text/template/parse"
)

// basePathField is the template expression a theme roots its own asset links
// with. It is named here because the check below quotes it in what it reports.
const basePathField = ".Site.BasePath"

// rootedAsset matches an HTML attribute whose value starts at the root of the
// domain, which is the one shape a theme must never write by hand: cress emits
// every URL it knows about already rooted, and the links a theme invents are
// the ones it does not know about.
//
// The two quoting styles are separate alternatives because RE2 has no
// backreference to match the closing quote to the opening one.
var rootedAsset = regexp.MustCompile(`(?i)\b(href|src|srcset|poster|action)\s*=\s*(?:"(/[^"]*)"|'(/[^']*)')`)

// protocolRelative starts a URL that names another host rather than this site,
// so no base path applies to it.
const protocolRelative = "//"

// Finding is one problem with a theme. Line is 1-indexed, and zero when the
// finding is about the theme as a whole rather than a place in a file.
type Finding struct {
	File    string
	Line    int
	Message string
}

// Report is what validating one theme found. A report with no findings is a
// theme that satisfies the contract as far as cress can tell without content.
type Report struct {
	// Name is the theme as cress.toml would name it.
	Name string
	// Path is where the theme was found, for a reader who wants to open it.
	Path     string
	Findings []Finding
}

// OK reports whether the theme passed.
func (r *Report) OK() bool { return len(r.Findings) == 0 }

// ValidateOptions names the theme to check and what to check it against.
type ValidateOptions struct {
	// SiteRoot is the site the theme is resolved from, the same way a build
	// resolves it.
	SiteRoot string
	// ThemesDir is the directory themes live in, relative to SiteRoot.
	ThemesDir string
	// Name is the theme's directory name.
	Name string
	// CressVersion is what the theme's declared contract is checked against.
	// Empty skips that check, which is what an untagged build wants.
	CressVersion string
}

// Validate checks a theme against the template contract and reports what is
// wrong, without building a site.
//
// A theme that cannot be found is an error, because there is nothing to
// validate. Everything about the theme itself is a finding instead, including
// the failures that would stop a build: the question asked here is what is
// wrong with this theme, and answering half of it as an error would report the
// first problem and hide the rest.
func Validate(opts ValidateOptions) (*Report, error) {
	fsys, themePath, err := locate(opts.SiteRoot, opts.ThemesDir, opts.Name)
	if err != nil {
		return nil, err
	}
	report := &Report{Name: opts.Name, Path: themePath}

	files, findings := collectTemplates(fsys)
	report.Findings = append(report.Findings, findings...)
	report.Findings = append(report.Findings, checkEntry(files)...)
	report.Findings = append(report.Findings, checkDuplicateDefines(files)...)
	report.Findings = append(report.Findings, checkEmptyLayouts(files)...)
	report.Findings = append(report.Findings, checkAssetLinks(files)...)
	report.Findings = append(report.Findings, checkFields(files)...)

	meta, err := loadMeta(fsys)
	if err != nil {
		report.Findings = append(report.Findings, Finding{File: MetaFile, Message: err.Error()})
	}
	for _, warning := range meta.Warnings(opts.CressVersion) {
		report.Findings = append(report.Findings, Finding{File: MetaFile, Message: warning})
	}

	sortFindings(report.Findings)
	return report, nil
}

// themeFile is one of a theme's template files, parsed on its own.
//
// Parsing each file separately is what makes a duplicate {{ define }} name
// visible: the loader parses them into one set, where a second definition
// replaces the first and leaves nothing to compare.
type themeFile struct {
	path   string
	layout bool
	text   string
	// trees holds the file's own body, keyed by path, plus a tree per
	// {{ define }} in it, keyed by the name it defined.
	trees map[string]*parse.Tree
}

// collectTemplates reads and parses every template in the theme, reporting the
// ones that do not parse.
//
// A file that fails to parse stays in the list with no trees. It is still one
// of the theme's files, and dropping it would have page.html reported missing
// whenever it merely had a syntax error, which is a second finding that sends
// the reader looking for the wrong thing. The checks that read trees see an
// empty set and say nothing; the ones that read the text still work.
func collectTemplates(fsys fs.FS) ([]*themeFile, []Finding) {
	files, findings := templatePaths(fsys)
	for _, file := range files {
		data, err := fs.ReadFile(fsys, file.path)
		if err != nil {
			findings = append(findings, Finding{File: file.path, Message: fmt.Sprintf("cannot be read: %v", err)})
			continue
		}
		file.text = string(data)

		trees, err := parseTrees(file.path, file.text)
		if err != nil {
			findings = append(findings, Finding{File: file.path, Message: parseMessage(err)})
			continue
		}
		file.trees = trees
	}
	return files, findings
}

// templatePaths lists a theme's layouts and partials, applying the same rules
// the loader does: a layout is a file directly in the layouts directory, and a
// partial is anything under partials/, however deeply nested.
func templatePaths(fsys fs.FS) ([]*themeFile, []Finding) {
	var files []*themeFile
	layouts, err := fs.Glob(fsys, layoutsGlob)
	if err != nil {
		return nil, []Finding{{Message: fmt.Sprintf("listing layouts: %v", err)}}
	}
	for _, layout := range layouts {
		files = append(files, &themeFile{path: layout, layout: true})
	}

	if info, err := fs.Stat(fsys, partialsDir); err != nil || !info.IsDir() {
		return files, nil
	}
	var findings []Finding
	err = fs.WalkDir(fsys, partialsDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.EqualFold(path.Ext(p), htmlExt) {
			files = append(files, &themeFile{path: p})
		}
		return nil
	})
	if err != nil {
		findings = append(findings, Finding{File: partialsDir, Message: fmt.Sprintf("cannot be read: %v", err)})
	}
	return files, findings
}

// parseTrees parses one file on its own, with function checking off. A theme
// may call any function cress registers, and a validator that resolved them
// would be a second list to keep in step with the renderer for no gain: an
// unknown function is a parse error at build time either way.
func parseTrees(name, text string) (map[string]*parse.Tree, error) {
	tree := parse.New(name)
	tree.Mode = parse.SkipFuncCheck
	trees := map[string]*parse.Tree{}
	if _, err := tree.Parse(text, "", "", trees); err != nil {
		return nil, err
	}
	return trees, nil
}

// parseMessage strips the layers text/template wraps a parse error in, which
// name the template package twice before saying what is wrong.
func parseMessage(err error) string {
	message := err.Error()
	for {
		trimmed := strings.TrimPrefix(message, "template: ")
		if trimmed == message {
			return message
		}
		message = trimmed
	}
}

// checkEntry reports a theme with no entry template. It is the one file a theme
// must have, and a theme missing it cannot render a page at all.
func checkEntry(files []*themeFile) []Finding {
	for _, file := range files {
		if file.layout && path.Base(file.path) == entryTemplate {
			return nil
		}
	}
	return []Finding{{Message: fmt.Sprintf("no %s: every theme needs one, since it renders any page that names no layout", path.Join("templates", entryTemplate))}}
}

// checkDuplicateDefines reports one {{ define }} name claimed by two files.
//
// Nothing complains about this at build time: every template parses into one
// set, so the file parsed last wins and the other definition is simply gone.
// Neither is more correct than the other, which is why this is reported rather
// than resolved.
func checkDuplicateDefines(files []*themeFile) []Finding {
	sources := map[string][]string{}
	for _, file := range files {
		for name := range file.trees {
			if name != file.path {
				sources[name] = append(sources[name], file.path)
			}
		}
	}

	var findings []Finding
	for name, paths := range sources {
		if len(paths) < 2 {
			continue
		}
		sort.Strings(paths)
		for _, p := range paths[1:] {
			findings = append(findings, Finding{
				File:    p,
				Line:    defineLine(files, p, name),
				Message: fmt.Sprintf("defines %q, which %s also defines; whichever parses last wins and nothing warns", name, paths[0]),
			})
		}
	}
	return findings
}

// checkEmptyLayouts reports a layout that renders nothing of its own. A file
// directly in the layouts directory is something a page can ask for by name, so
// one holding only {{ define }} blocks is a fragment in the wrong place: a page
// naming it renders blank, and the build says nothing.
func checkEmptyLayouts(files []*themeFile) []Finding {
	var findings []Finding
	for _, file := range files {
		if !file.layout {
			continue
		}
		if body, ok := file.trees[file.path]; ok && rendersNothing(body) {
			findings = append(findings, Finding{
				File:    file.path,
				Message: "is a layout but renders nothing outside its {{ define }} blocks; a fragment belongs under " + partialsDir,
			})
		}
	}
	return findings
}

// rendersNothing reports whether a template body is only whitespace once its
// {{ define }} blocks are taken out, which is how they are already stored.
func rendersNothing(body *parse.Tree) bool {
	if body.Root == nil {
		return true
	}
	for _, node := range body.Root.Nodes {
		text, ok := node.(*parse.TextNode)
		if !ok || strings.TrimSpace(string(text.Text)) != "" {
			return false
		}
	}
	return true
}

// checkAssetLinks reports a theme's own asset link that is not rooted under the
// site's base path.
//
// This is the subdirectory trap, and it is invisible in the only place people
// look: the link works at a domain root, so a theme is written, tested, and
// published before anybody publishes a site under a path and finds it renders
// unstyled. Cress cannot catch it at build time, because it cannot tell which
// URLs in a template the theme meant as its own.
func checkAssetLinks(files []*themeFile) []Finding {
	var findings []Finding
	for _, file := range files {
		for _, match := range rootedAsset.FindAllStringSubmatchIndex(file.text, -1) {
			attribute := group(file.text, match, 1)
			url := group(file.text, match, 2) + group(file.text, match, 3)
			if strings.HasPrefix(url, protocolRelative) {
				continue
			}
			findings = append(findings, Finding{
				File:    file.path,
				Line:    lineAt(file.text, match[0]),
				Message: fmt.Sprintf("%s=%q is not rooted under {{ %s }}, so it breaks when the site is published to a subdirectory", attribute, url, basePathField),
			})
		}
	}
	return findings
}

// group returns capture group n of a submatch, or empty when that alternative
// did not participate in the match.
func group(text string, match []int, n int) string {
	if len(match) <= 2*n+1 || match[2*n] < 0 {
		return ""
	}
	return text[match[2*n]:match[2*n+1]]
}

// defineLine finds where a file defines name, so a duplicate can be pointed at
// rather than described.
func defineLine(files []*themeFile, filePath, name string) int {
	for _, file := range files {
		if file.path != filePath {
			continue
		}
		if tree, ok := file.trees[name]; ok && tree.Root != nil {
			return lineOf(tree, tree.Root)
		}
	}
	return 0
}

// lineOf reports the 1-indexed line a node sits on. The parser tracks positions
// for its own error messages, and ErrorContext is the only way out to them.
func lineOf(tree *parse.Tree, node parse.Node) int {
	if tree == nil || node == nil {
		return 0
	}
	location, _ := tree.ErrorContext(node)
	// ErrorContext returns "name:line:col", and a template name may itself hold
	// colons, so the line is found from the right.
	parts := strings.Split(location, ":")
	if len(parts) < 3 {
		return 0
	}
	line := 0
	if _, err := fmt.Sscanf(parts[len(parts)-2], "%d", &line); err != nil {
		return 0
	}
	return line
}

// lineAt reports the 1-indexed line byte offset falls on.
func lineAt(text string, offset int) int {
	if offset > len(text) {
		offset = len(text)
	}
	return 1 + strings.Count(text[:offset], "\n")
}

// sortFindings puts a report in the order somebody would read the theme, so the
// output is stable between runs and between platforms.
func sortFindings(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].File != findings[j].File {
			return findings[i].File < findings[j].File
		}
		if findings[i].Line != findings[j].Line {
			return findings[i].Line < findings[j].Line
		}
		return findings[i].Message < findings[j].Message
	})
}

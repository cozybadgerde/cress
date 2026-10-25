# Software design audit: whole project

**Date:** 2026-10-25
**Scope:** Whole project, including uncommitted implementation and configuration.
**Overall rating:** 8/10 - Clear, acyclic package boundaries and a small,
appropriate pipeline; resource ownership has a few avoidable coupling points.

The baseline commit is `86aaf05`; findings refer to the working tree, not
only that commit. This is a design audit, not a security or release-readiness
assessment. No implementation changes are included.

## Summary

| # | Category | Finding | Severity (1-10) | Location | Effort | Status |
|---|----------|---------|-----------------|----------|--------|--------|
| 1 | Architecture | Writer combines site-lock ownership with a mutable output target | 4 | `internal/output/session.go:185-218` | M | Open |
| 2 | SOLID | Preview replaces a resource before acquiring its replacement | 3 | `internal/preview/preview.go:159-170` | S | Open |
| 3 | Architecture | Theme loading reads private files into the generation snapshot | 3 | `internal/theme/theme.go:115-173` | M | Open |
| 4 | Design Patterns | Scaffold specialization depends on unchecked text matches | 2 | `internal/scaffold/scaffold.go:79-81`, `internal/theme/init.go:77-80` | S | Open |

Severity reflects current impact, not hypothetical enterprise requirements.
Effort: S is under an hour; M is about half a day; L is over a day.
These are estimates including focused regression coverage.

## Findings

### #1 - Writer combines site-lock ownership with a mutable output target

**Category:** Architecture
**Severity:** 4/10
**Location:** `internal/output/writer.go:48-55`,
`internal/output/session.go:185-218`
**Effort:** M

**Why it's an issue:**

`Writer` represents two distinct lifetimes: the shared source site's writer
lock and a target's transaction state and filesystem. Normal builds use
the site root for both. Preview acquires that writer, closes its root,
replaces the root with a session directory, and rebinds its filesystem
callbacks while deliberately keeping the original site's lock.

The behavior is intentional and is covered by integration tests. The
concern is the mutable construction protocol: correct isolation depends on
updating several fields together before any transaction operation runs.
The type itself does not express which root the lock protects or which
target its operations affect.

**Where it occurs:**

`Session.Attempt` mutates `writer.root`, `writer.rename`, and
`writer.removeAll` at lines 199-203. Other session operations construct
partial `Writer` values just to use record-reading or verification methods
at lines 152, 247, 286, and 333.

Those values currently call compatible subsets of methods, but future
changes to `Writer` initialization would need to preserve these implicit
subsets across several call sites.

**Proposed fix:**

Keep site-lock lifetime separate from target transaction construction
inside `internal/output`. Construct the target-bound writer once, with
its root and default filesystem operations already consistent. Normal
builds and previews should share that constructor rather than rebinding an
acquired writer.

Start with private concrete helpers or types; no public writer interface,
filesystem framework, or alternate storage backend is needed.
Preserve the current site-wide lock duration and all recovery rules.
Retain the contention, target-isolation, recovery, and cleanup tests when
making this change.

### #2 - Preview replaces a resource before acquiring its replacement

**Category:** SOLID
**Severity:** 3/10
**Location:** `internal/preview/preview.go:87-93`,
`internal/preview/preview.go:159-170`
**Effort:** S

**Why it's an issue:**

The rebuild loop assigns the result of `os.OpenRoot` directly to
`pages.root`. If opening the newly published directory fails, the assignment
sets `pages.root` to nil. The old root is only closed when opening succeeds.
Returning on failure therefore bypasses explicit closure of the old root:
the deferred cleanup inspects the now-nil `pages.root`, not the old handle.

This is a concrete failure-path cleanup gap, not evidence that ordinary
rebuilds leak resources. Garbage collection or process exit may reclaim the
handle, but deterministic ownership is lost. The path was confirmed by
source inspection, not a fault-injected reproduction.

The SRP concern is local: resource replacement, page-server state,
publication-token handling, retry scheduling, and terminal reporting all
live inside the same rebuild branch. That makes lifecycle ordering harder
to check.

**Where it occurs:**

`run` saves `old := pages.root`, overwrites `pages.root` at line 162, then
returns at line 168 without closing `old` when `OpenRoot` fails.

**Proposed fix:**

Acquire the replacement in a local variable before updating the owned
field, while retaining the existing mutex and fatal-error behavior:

```go
nextRoot, err := os.OpenRoot(filepath.Join(session.Directory(), "public"))
if err != nil {
    pages.mu.Unlock()
    return err
}
old := pages.root
pages.root = nextRoot
if err := old.Close(); err != nil {
    pages.mu.Unlock()
    return err
}
pages.prefix, token = next.Prefix, next.Token
```

The existing deferred cleanup then retains an owned root on either branch.
If this logic grows further, extract only the root-swap operation rather
than introducing a preview service hierarchy.
Add focused failure-path coverage for acquiring the replacement root.

### #3 - Theme loading reads private files into the generation snapshot

**Category:** Architecture
**Severity:** 3/10
**Location:** `internal/theme/theme.go:115-173`
**Effort:** M

**Why it's an issue:**

`loadPackage` reads every regular file beneath the selected theme into a
byte map before selecting templates, public assets, and the active locale.
Private source files, design originals, other locales, or a nested Git
directory therefore contribute to read volume and temporary memory use
even though their contents cannot affect the generated page.

This is an architectural cost visible directly in the loader, not a
measured performance regression. It is modest for the bundled themes but
avoidable for locally developed themes containing larger private inputs.
Every preview rebuild repeats the work.

**Where it occurs:**

The walk calls `fs.ReadFile` and retains the result at lines 129-133.
Only afterward do lines 138-173 select `theme.json`, `templates/*.gohtml`,
`public/` files, and `locales/<language>.toml`.

**Proposed fix:**

Retain whole-package entry validation, including symlink and special-file
rejection, but retain file contents only for generation inputs.
Preserve the existing error behavior for unreadable private files, for
example by opening and closing them without loading their complete contents
when readability needs validation.

Use a small private path predicate and the existing `fs.FS` boundary, not
a repository abstraction or cache. An instrumented in-memory filesystem
test can verify that private file bodies are not read while required
inputs, diagnostics, ordering, and publication behavior remain unchanged.

### #4 - Scaffold specialization depends on unchecked text matches

**Category:** Design Patterns
**Severity:** 2/10
**Location:** `internal/scaffold/scaffold.go:79-81`,
`internal/theme/init.go:77-80`
**Effort:** S

**Why it's an issue:**

Site initialization changes the default theme through an exact text
replacement. Theme initialization similarly replaces a template block
with a stylesheet link. Both replacements silently leave the input
unchanged when the expected anchor disappears.

A valid edit to the embedded TOML spelling or template whitespace could
break specialization without producing an initialization error. Existing
integration coverage catches the current selected-theme behavior, but
the transformation itself does not state or enforce its precondition.
There is no demonstrated failure with today's embedded sources.

**Where it occurs:**

`scaffold.Init` replaces the first literal `name = "poppy"`.
`theme.Init` replaces the first literal
`{{block "theme-styles" .}}{{end}}`.

**Proposed fix:**

For the current two fixed substitutions, check that exactly one expected
anchor exists before replacing it and return a contextual error otherwise.
This is smaller than adding a general scaffold builder or reserializing
the commented TOML file.

Treat these as explicit template parameters only if more substitutions
become necessary. Cover missing and duplicate anchors when extracting any
small substitution helper; do not build a general templating subsystem.

## Architecture assessment

Cress uses a modular monolith with a staged generation pipeline, not MVC
or a collection of services. The main data flow is:

```text
CLI -> generator -> content preparation -> prepared page data
                -> theme loading + renderer -> prepared output files
                -> output transaction -> public/

CLI -> preview -> generator -> isolated output session
              -> source watcher + loopback HTTP adapter
```

The actual imports form an acyclic graph. `internal/page` is the shared
presentation contract. The renderer imports `page` and `theme`, but not
content discovery, generator orchestration, or output mutation.
`content` imports `theme` for name validation; that is a narrow concrete
dependency, not sufficient reason to introduce another shared package.

The generator is a small application facade. Content preparation spans
several concerns but separates them into metadata, discovery, references,
Markdown, lists, and publication files. There is no clear God object or
spaghetti dependency structure. Output recovery is necessarily branch-heavy
because it checks explicit filesystem states; its size alone is not a
reason to fragment the state machine.

The runtime uses local filesystem and operating-system facilities, plus
a loopback HTTP server for preview. It does not fetch content or invoke
deployment services. GitHub publishing and signing integrations live in
release configuration, outside the generation runtime.

### Bottlenecks and deliberate tradeoffs

- The full output set is held in memory before staging. Large public
  assets contribute to peak memory, but rendered pages and asset byte
  slices are generally passed onward rather than copied at every layer.
  Keep this simple model unless real site sizes justify streaming.
- `prepareLists` scans all pages for each list
  (`internal/content/lists.go:70-118`). Cost grows with lists times pages,
  plus sorting. This is acceptable for the intended small sites; section
  indexing is an optional optimization, not a required redesign.
- Preview holds the page-server write mutex across an entire rebuild
  (`internal/preview/preview.go:147-172`). Requests can wait during parsing
  and rendering as well as publication. The documented design permits
  this; it protects the currently served directory during replacement.
  Do not narrow the lock without preserving filesystem lifetime safety.
- The source watcher rescans directory watches after structural changes
  (`internal/preview/watch.go:49-85`, `104-111`). It favors correctness and
  simplicity over incremental watch bookkeeping. No scaling problem was
  measured.

The rating is 8/10 because boundaries and dependency direction are strong,
while lock/target construction and preview resource transitions still rely
on procedural invariants. A higher score does not require more layers.

## Design patterns assessment

| Pattern | Evidence | Assessment |
|---------|----------|------------|
| Factory/constructor | `command.NewCommand`, `render.NewPackageRenderer`, `output.NewSession` | Appropriate fallible construction where files or templates are involved. No singleton or builder framework is needed. |
| Facade/application service | `generator.BuildVersion`, `generate` | Thin coordination over content, themes, rendering, and output. Keep it concrete. |
| Adapter | `preview.publishedFS`, `pageServer`, Goldmark parser/renderer extensions | Appropriate library and HTTP boundaries; trusted HTML conversion stays in content preparation. |
| Observer/event loop | `sourceWatcher.run`, preview's select loop | Suitable for native filesystem events, debouncing, and contention retries. Channels coalesce work rather than introducing a general event bus. |
| Strategy seam | Private `buildFunc` passed into preview `run` | A small function supports focused tests. A multi-method build interface would not improve this seam. |
| DTO/value representation | `page.Page`, `page.Site`, `page.Entry`, `output.File` | Prepared values keep themes away from discovery and routing. Mutable Go fields are acceptable inside these package contracts. |
| Explicit state machine | `state.validate`, `Writer.Write`, `Writer.recover` | Appropriate for recovery and output ownership. State objects or a chain-of-responsibility framework would obscure filesystem invariants. |
| Template-based generation | Embedded scaffold transformations | Lightweight overall, but unchecked anchors need the local guard in finding #4. This is not a GoF Template Method hierarchy. |

There is no useful missing repository, proxy, decorator hierarchy, global
singleton, or dependency-injection container. Common filesystem-copying
steps in site/theme initialization have different destination and rewrite
rules; consolidating the whole operations would add more coupling than it
removes.

## SOLID assessment

| Principle | Assessment |
|-----------|------------|
| SRP | Package responsibilities are clear. Findings #1 and #2 identify the resource-lifecycle seams worth tightening; they do not call for reorganizing the entire application. |
| OCP | Local theme packages extend presentation without generator changes. The three page-type branches and two built-in names represent a deliberately closed product contract; polymorphism would be unnecessary. |
| LSP | There is no inheritance hierarchy to audit. The filesystem and HTTP adapters implement library interfaces, with explicit errors for intentionally unavailable files. No additional substitution issue was identified in the reviewed paths. |
| ISP | Standard interfaces and the private build function are narrow. No oversized application interfaces were found. Concrete packages are not, by themselves, ISP violations. |
| DIP | Stable prepared data and `fs.FS` keep useful boundaries. Concrete orchestration and constructors are reasonable for one runtime implementation. Function-valued filesystem hooks enable failure tests without requiring a broad filesystem interface. |

## Evidence and limits

The audit inspected production Go packages, representative unit and
filesystem/HTTP integration tests, schemas, embedded scaffold/theme
structure, architecture and authoring contracts, and release/task
configuration. The package graph was checked with `go list`.

The existing `mise run check` and `mise run build` completed successfully.
Passing tests support the current behavior; they do not disprove the
failure-path and maintenance findings above.

No browser testing, native Linux execution, workload benchmark, actual
release publication, or dedicated security review was performed.
The existing accessibility and release-evidence leftovers remain separate.
Recommendations are open proposals, not authorization for implementation
or a claim that every suggestion must block a release.

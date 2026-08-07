# CLAUDE.md - cress

Project notes for AI-assisted work on cress. This file is **cress-specific
only**. Role, communication, the change workflow, testing categories
(`unit`/`integration`/`e2e`), the commit-message format, and the architecture
pushback rules all live in the global `~/.claude/CLAUDE.md` and are not repeated
here. The repo's `.gitmessage` is the commit template.

## Read the guides first

cress documents itself for the people who use it, and those documents are the
source of truth for how it behaves. Work from them the way any contributor
would, rather than from a summary kept here: a summary is one more thing that
goes stale, and a wrong one is worse than none.

| Guide | Answers |
|---------------------------|--------------------------------------------------|
| [`README.md`](./README.md) | What cress is, and what it deliberately is not. |
| [`docs/USER.md`](./docs/USER.md) | `cress.toml`, front matter, navigation, choosing a theme, the 404, and every command that builds or previews a site. |
| [`docs/PUBLISH.md`](./docs/PUBLISH.md) | Getting output onto a host: Pages, `base_url` per target, own server. |
| [`docs/THEME.md`](./docs/THEME.md) | Writing a theme: the theme directory, layouts and partials, the template data contract, and what a theme must honor. |
| [`docs/OPERATOR.md`](./docs/OPERATOR.md) | Installing, pinning, upgrading, removing cress. |
| [`docs/DEVELOPER.md`](./docs/DEVELOPER.md) | Project layout, the design decisions behind it, image asset rules, commands, and CI. |

Two consequences worth stating outright:

- **A behaviour change is not finished until its guide matches.** The guides are
  written against real behaviour, so a change that leaves one describing the old
  behaviour has produced a bug in the documentation.
- **Do not restate them here.** If something about cress needs writing down and
  a contributor would want it too, it belongs in the guide that owns that
  audience. This file is for what has no place there.

## Issue tracking

The [GitHub issue tracker](https://github.com/cozybadgerde/cress/issues) is the
source of truth for planned work, bugs, and feature requests. Track new issues
there, not in a file; cress keeps no `TODO.md` backlog.

- Use the `gh` CLI (GitHub) for issue operations; it authenticates against the
  signed-in GitHub account and infers the repo from the git remote. Check
  `gh issue <verb> --help` when unsure; `gh --version` prints the build.

### Reading issues

```sh
gh issue list                   # open issues (the default state)
gh issue list -s all            # -s open|closed|all
gh issue list -l enhancement    # filter by label
gh issue view <n>               # title and body
gh issue view <n> --comments    # its comments
```

### Closing issues

GitHub auto-closes an issue when a commit or merged PR that lands on the default
branch carries a closing reference. Use the commit footer for this:

```text
- resolves: #7
```

Once that lands on the default branch, the issue closes on its own. Do not run
`gh issue close` for work resolved this way. For a non-code closure (wontfix,
duplicate), close manually with a reason and comment:

```sh
gh issue close <n> -r "not planned" -c "<reason>"
```

### Labels

On GitHub, labels are **repo-scoped** (unlike Codeberg's org-level set), so each
repo owns its own. List the current labels before tagging:

```sh
gh label list
```

Reuse the standard set where it exists: `bug`, `enhancement`, `question`,
`duplicate`, `invalid`, `good first issue`, `help wanted`, and the
project-specific `contribution welcome` and `upstream`. Create a missing one
with `gh label create "<name>"` rather than inventing near-duplicates.

### Creating an issue with `gh`

`gh issue create` takes `--label` directly; pass the body from a file or inline,
and `--template <name>` fills from a `.github/ISSUE_TEMPLATE` form:

```sh
gh issue create --title "<title>" --body-file <file> --label enhancement
gh issue comment <n> --body-file <file>       # add a comment
```

Issue templates are inherited from the org-wide defaults in
[`cozybadgerde/.github`](https://github.com/cozybadgerde/.github/tree/trunk/.github/ISSUE_TEMPLATE),
not kept in this repo. For an enhancement, mirror that template: a
`### Problem or motivation`, `### Proposed solution`, and `### Alternatives
considered` section. Footer the body with where the idea came from when it helps
traceability.

## Roadmap notes

Out of scope for 1.0: internationalization (i18n), and automatic list or section
pages (see #23, a candidate for 2.0). The absence of list pages is a scope
decision for the initial release, not a permanent one.

Deliberately not planned at any version: plugins or a JS ecosystem, data files,
shortcodes, taxonomies, and pagination. A request for one of these is a request
to be a different generator, and the answer is the comparison table in the
README rather than an implementation.

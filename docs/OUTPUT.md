# Build and preview document

This document describes how preview and builds behave, how Cress protects
your output, and how to recover after an interruption. The
[user guide](USER.md) covers everyday use.

## Preview

```sh
cress site preview
cress site preview --drafts --port 0 my-site
```

`cress site preview --help` lists the default port and all flags.

### Serving

- Preview builds into its own private temporary folder and never needs or
  changes `public/`. Several previews can run at once.
- It listens only on `127.0.0.1`. `--port 0` picks a free port. A busy port
  fails instead of silently picking another.
- It doesn't open a browser and has no option for remote access.
- It serves below `site.path` and redirects `/` there. Canonical and sitemap
  URLs keep `site.host`. Changing `site.path` prints the new address.
- Preview content is as trusted as published content, including raw HTML and
  theme templates.

### Drafts

Drafts appear only with `--drafts`. The choice lasts for the whole session,
including draft assets and list entries. Shown drafts need a title and valid
settings. Normal builds never include drafts.

### Rebuilding

Preview watches `cress.toml`, `content/`, `static/`, and `themes/`, including
new folders and editors that save by replacing files. Changes to `public/`
and `.cress/` are ignored. Quick successive edits are combined, and edits
during a build trigger another build.

Watching relies on native file events, which network and virtual drives may
not provide. If watching fails, preview stops with an error rather than
missing edits.

The terminal reports each rebuild. Refresh the browser yourself. Preview
injects no reload script or error overlay.

- An invalid edit keeps the last working version visible.
- A rebuild blocked by a running `cress site build` retries about every half
  second.
- Other failures are reported and wait for the next edit.
- Requests may wait briefly while a rebuild finishes.
- If preview can't confirm a usable version, it stops and names its
  temporary folder.

The first build at startup must succeed, or preview doesn't start.

### Stopping

Ctrl-C or SIGTERM stops the server and removes its temporary output. On
startup, Cress cleans up leftovers from earlier previews of the same site that
are no longer running. It leaves running sessions and anything it doesn't
recognize alone. Unclear leftovers are kept and reported. Don't delete
session locks to force cleanup. The operating system may clear temporary
files on its own.

## Output protection

### Ownership

Cress writes only to a `public/` folder that is new, empty, or one it created
itself. It refuses other folders, symlinks, and files without changing them.
Each build replaces its output completely, including stale files and manual
additions.

Cress recognizes its output by pairing a public marker,
`public/.cress-output.json`, with private state in `.cress/output.json`. The
marker holds only a format version and a random token. Your content can't
create a file with that name.

- Keep `.cress/` out of version control and deployment. Only you can read it.
- Don't edit these records or remove the lock file in `.cress/`.
- Copying a deployed `public/` into another project doesn't make it that
  project's output.

Output from Cress versions before ownership support isn't recognized. Check
it, delete only the old `public/` once, and build again. There is no automatic
adoption or force option.

### One build at a time

Each build locks the site before reading anything and holds the lock until it
finishes. A second build fails immediately with a "busy" message, even
through a different path to the same site. The lock disappears when the
process ends. An existing lock file doesn't mean a build is running, so don't
delete it to get past the message. The lock protects against cooperating
local processes on Linux and macOS, not remote writers or manual edits.

### Replacing output

Cress builds the whole site into `.cress/stage-<token>/` before touching
`public/`. During the switch, the previous output moves to
`.cress/backup-<token>/`.

- Input, rendering, and staging failures leave the previous output as it was.
- If the switch fails, Cress restores the verified previous output where it
  can, and reports a failed restore too.
- If cleanup fails after a successful switch, the new output stays but the
  command still fails.

Visitors may notice a short interruption during the switch. Cress doesn't
promise to survive a power loss.

### Recovery

The next build finishes or cleans up an interrupted one when the saved state
matches the ownership tokens. When it doesn't, Cress stops, deletes nothing,
and names the folders it kept. A leftover `.cress/output.next`, a changed
marker, or an unknown stage folder needs your attention.

To recover by hand:

1. Stop all Cress processes for the site.
2. Copy the kept folders somewhere safe.
3. Compare the previous and current tokens in the transaction with the
   output markers.
4. Restore only a backup whose token matches.

Don't delete `.cress/` to get past an error. There is no repair command yet.

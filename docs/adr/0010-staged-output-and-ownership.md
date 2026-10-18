# Stage complete output and guard directory ownership

Status: accepted architectural direction; concrete mechanisms are
implementation choices.

Date: 2026-10-16.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Cress must remove stale generated files without installing incomplete
generated results, destroying previous successful output after a
generation error, or deleting unrelated data.

Here, publication means replacing local generated output, not deploying a
site to a host.

Publication can fail or be interrupted. Concurrent builds and temporary
previews add ownership and recovery concerns that cannot safely depend on
directory names or guessed process liveness.

## Decision drivers

- Preserve previous output when generation fails.
- Remove stale files without mixing old and new output.
- Protect unrelated directories and source files.
- Make ownership and competing-writer behavior explicit.
- Keep recovery and visibility guarantees honest across supported systems.
- Protect draft previews without accumulating a persistent preview cache.

## Considered options

Generation:

- Stage complete output and replace wholly generated results.
- Preserve manual additions through incremental output tracking.

Ownership:

- Verify Cress ownership before replacing existing nonempty output.
- Treat any directory at the expected output path as disposable.

Recovery:

- Support recoverable replacement with verified recovery.
- Require manual recovery after any interruption.
- Promise uninterrupted atomic visibility.

Writer coordination:

- Use OS-managed advisory locking with fail-fast writer attempts.
- Make all competing commands wait.
- Infer ownership from PID files or process checks.

Preview retention:

- Clean temporary output on shutdown and safely reclaim verified orphans.
- Require manual cleanup after interrupted previews.

## Decision outcome

Chosen generation option: "Stage complete output and replace wholly
generated results", because it removes stale files without mixing
previous and newly generated content.

Generate and validate the complete result before replacing successful
output. Validation, rendering, and copying failures leave previous output
untouched. Manual additions to recognized generated output are not
preserved; public inputs belong in the source project.

Chosen ownership option: "Verify Cress ownership before replacing existing
nonempty output", because an expected path name does not make data
disposable.

Allow new or empty output destinations, but refuse unrecognized nonempty
directories, symlinks, and non-directory targets without modifying them.
Ownership evidence must not expose secrets or collide with user output.
Do not silently adopt unrelated output.

### Recoverability

Chosen recovery option: "Support recoverable replacement with verified
recovery", because interrupted publication must retain a safe path to
usable output without promising universally atomic replacement.

If publication fails, restore previous output where possible.
If restoration fails, preserve recoverable content and evidence and
report both failures. Do not claim success or guess destructive recovery
from names alone.

Recovery proceeds only when ownership and state can be verified.
Ambiguous states stop with actionable guidance rather than speculative
promotion or deletion.

Cleanup failure after confirmed publication leaves valid new output in
place and reports an explicit unsuccessful outcome. Do not roll back a
confirmed publication merely to hide cleanup errors.

Replacement may interrupt readers; it is not guaranteed to look like one
instantaneous switch. Cress does not guarantee that output or recovery
state survives power loss.
The recorded state, filesystem operations, and recovery steps are
implementation choices constrained by these guarantees.

### Writer coordination

Chosen coordination option: "Use OS-managed advisory locking with
fail-fast writer attempts", because cooperating Cress processes need
exclusive writer ownership without stale process-liveness guesses.

Coordinate output mutation and recovery per site, including preview
generation. Each writer attempt fails clearly on contention rather than
proceeding unprotected. Explicit builds and initial preview startup fail
fast. Established preview sessions retain blocked automatic rebuilds and
retry after contention clears, without requiring another edit or restart.
Each running session must catch up to current source when generation can
succeed. Report contention and pending retries without claiming success.
Other coordination failures must also be explicit and do not qualify for
contention retries. Different names or path aliases for one site must not
bypass exclusivity.

Read-only preview serving does not hold writer ownership for its lifetime.
PID files and process age are not proof of writer ownership.
Distributed locking and protection from manual edits are not promised.
Exact lock placement, lifetime management, and scheduling belong to the
implementation.

### Temporary isolated previews

Each preview session owns private temporary output, separate from
deployable output and other sessions. Draft visibility cannot leak between
sessions or into normal builds.
Normal builds never promote preview artifacts into deployable output.

Use the normal complete generation pipeline and output-protection
boundaries for each preview target, including automatic rebuilds.
Preserve the session's target and draft mode. Incremental rendering is not
required. Preview recovery cannot replace deployable output.

Keep established previews serving verified usable output and watching for
edits after recoverable publication or cleanup failures. This is restored
previous output after failed publication, or confirmed new output after
cleanup failure; report the rebuild as unsuccessful in either case.
If publication or recovery cannot establish verified usable output, stop
serving and exit unsuccessfully with actionable recovery guidance.
Preserve recoverable content and evidence rather than attempting
speculative repairs.

Changes arriving during rebuilding must not create competing writers for
the same session. Publication, recovery, and writer-contention failures
must remain explicit and must not bypass output protection.
Watcher configuration and event scheduling are implementation choices.

Chosen retention option: "Clean temporary output on shutdown and safely
reclaim verified orphans", because previews should not become a persistent
cache or be deleted on guessed liveness.

Protect preview state with appropriate access restrictions. Clean the
session's output on orderly shutdown and reclaim verified inactive
sessions after interruption. Preserve active sessions, unrelated files,
and ambiguous recovery evidence.

Report cleanup failures. Concrete session identity, ownership checks,
state layouts, and cleanup procedures are implementation choices.

### Consequences

- Good, because generation errors preserve previous successful output.
- Good, because complete replacement removes stale files.
- Good, because explicit ownership protects unrelated data.
- Good, because verified recovery avoids guessed destructive actions.
- Good, because active previews and their draft modes remain isolated.
- Bad, because staging and recovery need additional storage and state.
- Bad, because manual additions to generated output are discarded.
- Bad, because readers may observe a publication interruption.
- Bad, because interrupted previews can retain files until safe cleanup.

## Confirmation

When output handling is implemented, check complete replacement, stale-file
removal, preservation after generation errors, and ownership protection.
Exercise publication and interruption failures, competing writers, preview
isolation, and safe cleanup against the documented guarantees.
Check recoverable failures keep established previews running, unverifiable
output stops them, and competing sessions automatically catch up while
explicit builds and initial preview startup remain fail-fast.
These are implementation validation goals, not a prescribed transaction
state machine or evidence of completed functionality.

## More information

See [the product requirements](../PRD.md),
[content and asset boundaries](0003-content-tree-and-routes.md), and
[CLI command organization](0009-workflow-cli.md).

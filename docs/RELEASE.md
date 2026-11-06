# Release maintenance

## Generated reports

GoReleaser invokes [go-licenses](https://github.com/google/go-licenses)
to generate a CSV dependency-license report for each supported platform.
Each row contains a library path, license URL, and detected license name,
without a header. These `.licenses.csv` files are release artifacts,
covered by the signed checksum manifest.

The before hook also uses `go-licenses report` with
[the notices template](../.github/notices.tpl) to generate license texts
for the archives.
`mise run check:licenses`, included in `mise run check`, uses
`go-licenses check` to enforce the approved dependency-license allowlist for
all three release targets. Unknown or unapproved licenses fail the check.
Approval applies to license types, not individual dependency records; see
the [license policy](adr/0011-bsd-3-clause-license.md#dependency-license-policy).
Bundled non-code assets still need separate distribution-rights review.
The approved names include OFL-1.1 for font software; Chroma's combined
COPYING file causes detection even though its SVG font formatter is not
imported. Generated notices retain the font copyright and license text.
There is no custom release executable, exclusion list, or checked-in
generated inventory.

To print a report locally:

```sh
mise exec -- go-licenses report ./cmd/cress
```

## Local release rehearsal

```sh
mise run check
mise run build
mise run check:goreleaser
mise exec -- goreleaser release --snapshot --clean --skip=publish,sign
mise run test:release
```

The snapshot command replaces generated `dist/` artifacts and writes
`bin/THIRD-PARTY-NOTICES.txt`. It does not tag, publish, or sign.
The [release configuration](../.goreleaser.yml) owns target platforms,
archive contents, and report names.

The smoke test checks archive contents, binary targets, license reports,
and archive/report checksums. It extracts only the host-compatible executable
into a temporary directory, verifies its version, and initializes and builds
sites with all three embedded themes. Other targets are inspected, not run.
This does not replace native-platform acceptance.
The test never modifies `dist/` or existing sites. Ordinary checks skip it
unless `CRESS_RELEASE_DIR` is set. The explicit task defaults to `dist/`
and fails if artifacts are missing.

## Tag workflow

The [release workflow](../.github/workflows/release.yml) validates full
`v`-prefixed semantic version tags. A read-only job runs the existing checks
and build.
The publication job runs GoReleaser with release-write and OIDC permissions.
It refuses to start if a release or draft already exists for the tag. It
uploads signed artifacts to an unpublished draft, runs the archive smoke
test, and compares the uploaded asset digests with the tested local files.
It then publishes that exact draft by ID, preserving its prerelease flag.
Nothing is rebuilt between testing and publication.

A failure can leave one or more drafts. Do not manually publish them or
rerun blindly. Inspect the failed run and all drafts for that tag, resolve
the cause, and explicitly remove the failed drafts before retrying. Confirm
that each draft is the intended failed draft before removing it. The next
run refuses to proceed while any release or draft for that tag remains.
Tools use the versions pinned in [`mise.toml`](../mise.toml).

Choose the intended version rather than copying a documentation example.
Use a full major, minor, and patch version after `v`, without an intervening
hyphen. The workflow owns the exact tag-validation rules.
GoReleaser marks versions with a prerelease suffix as GitHub prereleases.
Build metadata alone does not make a release a prerelease.
Publishing the draft preserves its prerelease flag. A prerelease phase is
optional: publish the intended prerelease tag, assess it, then use a separate
stable tag when ready. Each tag follows the same automated draft gate.

GoReleaser injects the executable version, creates the archives and CSV
reports, and generates `checksums.txt`. Cosign keylessly signs that file.
The expected certificate identity is:

```text
https://github.com/cozybadgerde/cress/.github/workflows/release.yml@refs/tags/<exact-tag>
```

The issuer is `https://token.actions.githubusercontent.com`.
See [installation and verification](INSTALL.md) for download verification.

Tagging and publishing require the maintainer's approval.

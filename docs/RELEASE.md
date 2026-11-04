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
```

The snapshot command replaces generated `dist/` artifacts and writes
`bin/THIRD-PARTY-NOTICES.txt`. It does not tag, publish, or sign.
The [release configuration](../.goreleaser.yml) owns target platforms,
archive contents, and report names.

## Tag workflow

The [release workflow](../.github/workflows/release.yml) validates full
`v`-prefixed semantic version tags. A read-only job runs the existing checks
and build.
The publication job runs GoReleaser with release-write and OIDC permissions.
Tools use the versions pinned in [`mise.toml`](../mise.toml).

Choose the intended version rather than copying a documentation example.
Use a full major, minor, and patch version after `v`, without an intervening
hyphen. The workflow owns the exact tag-validation rules.
GoReleaser marks versions with a prerelease suffix as GitHub prereleases.
Build metadata alone does not make a release a prerelease.

GoReleaser injects the executable version, creates the archives and CSV
reports, and generates `checksums.txt`. Cosign keylessly signs that file.
The expected certificate identity is:

```text
https://github.com/cozybadgerde/cress/.github/workflows/release.yml@refs/tags/<exact-tag>
```

The issuer is `https://token.actions.githubusercontent.com`.
See [installation and verification](INSTALL.md) for download verification.

Outstanding accessibility, native-platform, and signature-verification
work remains in the
[work plan](../TODO.md#phase-3-release-tooling-and-acceptance).
Tagging and publishing require the maintainer's approval.

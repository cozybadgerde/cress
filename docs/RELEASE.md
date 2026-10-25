# Release maintenance

## Generated reports

GoReleaser invokes [go-licenses](https://github.com/google/go-licenses)
to generate a CSV dependency-license report for each supported platform.
Each row contains a library path, license URL, and detected license name,
without a header. These `.licenses.csv` files are release artifacts,
covered by the signed checksum manifest.

The before hook also uses `go-licenses report` with
`.github/notices.tpl` to generate license texts for the archives.
There is no custom release executable, license policy check, exclusion
list, or checked-in generated inventory.

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
Archives cover macOS arm64, Linux amd64, and Linux arm64.

## Tag workflow

`.github/workflows/release.yml` triggers on full `v`-prefixed semantic
version tags. A read-only job runs the existing checks and build.
The publication job runs GoReleaser with release-write and OIDC permissions.
Tools use the versions pinned in [`mise.toml`](../mise.toml).

Use `v1.0.0` for a stable release or `v1.0.0-pre.0` (also `-alpha.1`,
`-beta.1`, or `-rc.1`) for a prerelease. GoReleaser automatically marks
versions with a prerelease suffix as GitHub prereleases, not stable releases.
There is no hyphen between `v` and the major version:
`v-1.0.0-pre.0` is invalid. Short versions such as `v1.0` and numeric
components with leading zeros, such as `v01.0.0` or `v1.0.0-pre.01`,
are rejected. Optional SemVer build metadata, such as `+build.1`, is allowed;
build metadata alone does not make a release a prerelease.

GoReleaser injects the executable version, creates the archives and CSV
reports, and generates `checksums.txt`. Cosign keylessly signs that file.
The expected certificate identity is:

```text
https://github.com/cozybadgerde/cress/.github/workflows/release.yml@refs/tags/<exact-tag>
```

The issuer is `https://token.actions.githubusercontent.com`.
See [installation and verification](INSTALL.md) for download verification.

No release has been published yet. Outstanding accessibility, native
platform, and published-signature evidence remains in the
[work plan](../TODO.md#phase-9-documentation-and-release).
Tagging and publishing require the maintainer's approval.

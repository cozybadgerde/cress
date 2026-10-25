# Install and verify Cress

No release has been published yet. These instructions describe the
configured release artifacts. For a source build, use the developer guide
in the repository.

## Select an archive

Download from <https://github.com/cozybadgerde/cress/releases>.
Choose `darwin_arm64` for Apple Silicon macOS, `linux_arm64` for Linux
ARM64, or `linux_amd64` for Linux Intel/AMD 64-bit. Windows, Intel macOS,
and 32-bit ARM are not supported targets.

Download the archive, the matching platform's `.licenses.csv`, `checksums.txt`,
and `checksums.txt.sigstore.json` from the same release.
The executable embeds its themes and needs no Go
toolchain or separate runtime theme files.

## Verify before extracting

Install Cosign v3 using <https://docs.sigstore.dev/cosign/system_config/installation/>.
Use a shell in the download directory and set the exact downloaded tag:

```sh
tag=v1.0.0
cosign verify-blob \
  --certificate-identity "https://github.com/cozybadgerde/cress/.github/workflows/release.yml@refs/tags/$tag" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --bundle checksums.txt.sigstore.json checksums.txt
```

`v1.0.0` is an example, not an existing release. Never weaken the expected
identity or issuer to make verification pass. A valid signature from an
unexpected publisher is not sufficient.

Only after signature verification succeeds, check the archive and report
digests. With GNU coreutils:

```sh
sha256sum --check --ignore-missing checksums.txt
```

Confirm both downloaded filenames appear with `OK`. On macOS, the bundled
Perl utility can verify the manifest instead:

```sh
shasum -a 256 --check checksums.txt
```

Missing artifacts cause that command to fail; inspect every result and
require `OK` for both downloaded files. Do not treat an unlisted file as
verified. Stop on any bad signature, wrong identity, or digest mismatch.
The bundle packages verification evidence; trust still depends on the
expected identity, issuer, and Sigstore trust roots.

## Install

Extract only the verified archive. For example, on Linux amd64:

```sh
tar -xzf cress_1.0.0_linux_amd64.tar.gz
mkdir -p "$HOME/.local/bin"
install -m 755 cress "$HOME/.local/bin/cress"
```

Replace the example archive name with the verified download. Add
`$HOME/.local/bin` to your shell's `PATH` if needed. Keep `LICENSE`, the
third-party notices, and this reference when redistributing the archive.

```sh
cress --version
cress site init my-site
cress site preview --port 0 my-site
```

Open the printed URL; stop with Ctrl-C. Edit the example content and set
`site.host` before publishing. `cress site build my-site` produces
`my-site/public/`; deploy those files with external tooling.

macOS archives are not Apple Developer ID signed or notarized. Cosign
verification is not Apple notarization; macOS may display trust warnings.
Do not disable system-wide protections to install Cress.

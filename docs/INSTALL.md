# Installation guide

Check the [releases page](https://github.com/cozybadgerde/cress/releases)
for available archives. If no suitable release is available, follow
[local source setup](DEVELOPER.md#local-setup) and
[build the executable](DEVELOPER.md#building). The source build writes
`bin/cress`. Add that directory to your `PATH` or use the executable's full
path in place of `cress` in the guides.

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

Install Cosign using
<https://docs.sigstore.dev/cosign/system_config/installation/>.
Use a shell in the download directory. Enter the exact release tag,
including its leading `v`, at the prompt:

```sh
printf 'Release tag: '
IFS= read -r tag
cosign verify-blob \
  --certificate-identity "https://github.com/cozybadgerde/cress/.github/workflows/release.yml@refs/tags/$tag" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --bundle checksums.txt.sigstore.json checksums.txt
```

Use the tag of the release you downloaded, not a branch or a different
release. Never weaken the expected identity or issuer to make verification
pass. A valid signature from an unexpected publisher is not sufficient.

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

Extract only the verified archive. Enter its exact filename at the prompt:

```sh
printf 'Verified archive filename: '
IFS= read -r archive
tar -xzf "$archive"
mkdir -p "$HOME/.local/bin"
install -m 755 cress "$HOME/.local/bin/cress"
```

Add `$HOME/.local/bin` to your shell's `PATH` if needed. Keep `LICENSE`,
`THIRD-PARTY-NOTICES.txt`, and this guide when redistributing the archive.

```sh
cress --version
```

Continue with [creating your site](USER.md#create-your-site).

macOS archives are not Apple Developer ID signed or notarized. Cosign
verification is not Apple notarization; macOS may display trust warnings.
Do not disable system-wide protections to install Cress.

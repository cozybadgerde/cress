# Distribute signed release archives with GoReleaser

Status: accepted.

The bundled-theme count is superseded by
[ADR 0020](0020-third-bundled-theme.md). Other decisions remain unchanged.

Date: 2026-10-06.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Cress needs an installable release for Linux build environments and
Linux/macOS development. Users need to verify downloaded artifacts without
requiring a package manager or container runtime.

Releasing the generator is separate from deploying a generated site.
The build-only product boundary does not prevent publishing Cress releases.

## Decision drivers

- Support Apple Silicon, Linux ARM64, and Linux Intel/AMD 64-bit systems.
- Automate releases with GoReleaser.
- Provide verifiable release authenticity and integrity.
- Keep installation independent of external runtime theme files.
- Defer extra distribution channels and Apple-specific signing operations.

## Considered options

- Downloadable archives with signed checksums through GitHub Releases.
- Archives plus Apple Developer ID signing and notarization at launch.
- Container images or Homebrew as primary launch distribution channels.
- Unsigned archives or checksums without authenticated signatures.

## Decision outcome

Chosen option: "Downloadable archives with signed checksums through GitHub
Releases", because it covers the initial platforms with a small distribution
surface and a documented verification path.

### Targets and contents

Use GoReleaser to produce downloadable archives for:

- macOS Apple Silicon.
- Linux ARM64.
- Linux Intel/AMD 64-bit.

ARM64 does not include 32-bit ARM. The amd64 target supports both Intel and
AMD processors. Intel macOS and Windows are outside the 1.0 release matrix.

Each archive contains:

- The `cress` executable.
- The BSD-3-Clause license and any required bundled third-party notices.
- A short installation and verification reference.

Embed both built-in themes in the executable. Users do not install a
separate theme directory to use a built-in theme.
Also embed the minimal renderable skeleton used for theme initialization.
It is developer scaffolding, not a third selectable built-in theme, and
requires no download or separate installation.
Third-party themes remain local site inputs under
[the theme integration decision](0007-theme-packages-and-integration.md).

### Automation and channel

Use a version-tag-triggered GitHub Actions workflow with GoReleaser to
publish artifacts to GitHub Releases.
Concrete artifact layout, workflow configuration, and platform support
details belong to the release implementation and its documentation.

Release automation may use network services. Offline operation applies to
site generation, not to downloading Cress or publishing its releases.

### Authenticity and integrity

Generate release-specific machine-readable SBOM artifacts for bundled
components under [the license policy](0011-bsd-3-clause-license.md).
Publish them alongside the release archives.

Authenticate release checksums with Cosign keyless signing and publish
the verification evidence. Cover both archives and release SBOM artifacts.

Verification must establish the expected release workflow identity and
issuer, not merely a cryptographically valid signature. Authenticate
checksums before trusting downloaded artifacts.
Document concrete verification instructions with the release pipeline.

A checksum alone detects changes but does not establish who released the
archive. Signing a checksum manifest authenticates the listed digests;
each archive does not need a separate signature.

Keyless signing avoids maintaining a persistent release signing key.
It introduces reliance on the signing identity and Sigstore services
during release. Verification tooling and trust requirements must be
documented. A verification bundle packages signing evidence; it does not
remove the need to check the expected publisher identity and issuer.

### Deferred distribution work

Cosign release verification is not Apple Developer ID code signing or
notarization. Defer both Apple mechanisms and document that macOS releases
are not Apple-notarized. Do not promise warning-free macOS installation.

Container images and Homebrew distribution remain future options.
They can reuse the executable with embedded themes without changing the
site build contract. No additional distribution channel is selected by
this document.

### Consequences

- Good, because three archives cover the agreed initial release targets.
- Good, because built-in themes need no separate runtime installation.
- Good, because authenticated checksums support artifact verification.
- Good, because keyless signing avoids a long-lived signing secret.
- Bad, because users must perform verification with additional tooling.
- Bad, because release signing depends on external identity services.
- Bad, because macOS trust checks are not addressed by Cosign signing.
- Bad, because embedded built-in themes update with the Cress executable.

## Confirmation

When releases are implemented, check supported-platform operation, offline
built-in rendering and theme initialization, and required archive contents.
Confirm authenticity and integrity verification covers archives and SBOMs,
rejects tampering and unexpected identities, and reflects license review.
Document concrete verification and macOS trust limitations with releases.

## More information

- [Product requirements](../PRD.md).
- [Implementation language](0001-go-implementation-language.md).
- [Generation and deployment boundary](0002-offline-generation-and-deployment-boundary.md).
- [Theme packages and integration](0007-theme-packages-and-integration.md).
- [Accessibility baseline and assessment](0016-accessibility-baseline-and-assessment.md).
- [GoReleaser artifact signing](https://goreleaser.com/customization/sign/sign/).
- [GoReleaser Cosign bundle guidance](https://goreleaser.com/blog/cosign-v3/).

This accepted decision guides release architecture. It does not claim
that release automation or signing has been implemented.

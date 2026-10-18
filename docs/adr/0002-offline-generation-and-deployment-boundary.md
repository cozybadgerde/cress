# Generate sites offline and keep deployment external

Status: accepted.

Date: 2026-10-06.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Users want to build websites locally or through continuous integration and
continuous delivery (CI/CD), then publish them to different hosting targets.
Cress needs a clear responsibility boundary between generating a website
and delivering it to a host.

Supporting the complete workflow raises the question of whether Cress
should also manage uploads, deployment protocols, and hosting integrations.

## Decision drivers

- Keep reliable website generation the product's primary focus.
- Keep ongoing maintenance manageable for a small company, accounting for
  deployment protocols, provider changes, and credential handling.
- Support local and CI/CD workflows across different hosting targets.
- Compose with specialized deployment tools rather than duplicate their
  responsibilities without a demonstrated need.
- Keep generation usable offline with locally available inputs.

## Considered options

- Generate local static output with deployment handled externally.
- Combine generation with built-in deployment through upload protocols,
  such as FTP or SCP, and hosting-provider integrations.

## Decision outcome

Chosen option: "Generate local static output with deployment handled
externally", because it keeps Cress focused on generation and avoids owning
a growing deployment compatibility and maintenance surface.

Protocol support and hosting integrations would require ongoing work
beyond generating websites. Specialized deployment tools already address
that responsibility and can be composed with Cress locally or in CI/CD.
Building another deployment manager does not provide enough product value
to justify that scope now.

Generation produces local static files. It does not upload, deploy, or
provision hosting. Themes and other build inputs must already be available
locally; generation does not download them.

Reconsider built-in deployment only if a concrete user need justifies the
expanded product scope and ongoing maintenance. Possible future demand is
not an initial release commitment.

This boundary does not prohibit downloading Cress or using network
services to publish Cress releases. It does not select the implementation
language, CPU targets, or release packaging.

### Consequences

- Good, because deployment credentials and provider APIs stay outside Cress.
- Good, because maintenance remains focused on website generation rather
  than a growing set of deployment targets.
- Good, because local and CI builds share the same generation boundary.
- Good, because maintainers can choose specialized deployment tools and
  change hosting targets independently of Cress.
- Bad, because maintainers need another mechanism to deploy sites.

## Confirmation

When generation is implemented, confirm representative sites build without
network access and produce local output without initiating deployment.

## More information

See [the implementation language](0001-go-implementation-language.md),
[release distribution](0012-release-artifacts.md), and
[the product requirements](../PRD.md).

# Use Go as the implementation language

Status: accepted.

Date: 2026-10-06.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Cress needs an implementation language for its generator and CLI.
Maintainers need reliable native executables on the supported Linux and
macOS platforms without installing or managing a separate language runtime.
The language must support site generation and CLI development while keeping
release artifacts straightforward to build and distribute.

## Decision drivers

- Support the agreed platforms with native executables.
- Distribute self-contained binaries without a separate interpreter or
  application runtime.
- Provide the features and libraries needed for a native CLI and generator.
- Have good documentation and a stable language and library ecosystem.
- Support predictable release artifacts and established release and
  security tooling.

## Considered options

- Implement Cress in Go.
- Implement Cress in JavaScript or TypeScript.
- Implement Cress in Python.

## Decision outcome

Chosen option: "Implement Cress in Go", because it best fits Cress's native
CLI and self-contained distribution needs across the supported platforms.

Go can compile the application and its Go dependencies into a native
executable.
Its documentation and ecosystem support CLI development, release packaging,
and security checks. Libraries such as Cobra and urfave/cli, and tools such
as GoReleaser and gosec, illustrate that fit.

JavaScript and TypeScript were considered, but producing similarly
self-contained CLI artifacts was judged to require more runtime, bundling,
and packaging work. Their ecosystem was a less direct fit for this
distribution goal, not incapable of supporting CLIs.

Python was considered, but relying on a separately installed interpreter
introduces variation across operating systems and OS versions. Bundling an
interpreter would add packaging work to avoid that dependency.

Preserve self-contained delivery when selecting dependencies and the build
approach. This means no separately installed application runtime, not an
absence of code dependencies.

This decision does not select a Go version, package layout, CLI framework,
or security tool. Concrete dependency choices belong to implementation.

Offline operation, deployment scope, rendering, and distribution are
separate architectural decisions.

### Consequences

- Good, because users do not manage interpreter or application-runtime
  versions to run Cress.
- Good, because the ecosystem supports native CLI development and release
  maintenance.
- Bad, because native artifacts must be built for each supported target.
- Bad, because dependencies that introduce external runtime requirements
  would undermine the distribution goal.

## Confirmation

When implemented, confirm the Go-built generator and CLI run on supported
targets without a separately installed interpreter, application runtime,
or Go toolchain.

## More information

See [the generation boundary](0002-offline-generation-and-deployment-boundary.md),
[release distribution](0012-release-artifacts.md), and
[the product requirements](../PRD.md).

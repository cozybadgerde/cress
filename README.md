# Cress

A cozy static site generator. Fresh little sites, fast.

Cress turns Markdown content, one TOML configuration file, and a curated
theme into static website files. It is built by
[Cozy Badger](https://www.cozybadger.de) for small businesses and
solopreneurs first.

## Motivation

Small businesses need a simple, reliable website. Cress focuses on exactly
that: a few page types, clear configuration, and themes that look good
without template knowledge.

Cress keeps the site structure small and predictable. A technical
maintainer runs the site, while the business owner only writes Markdown.
High-quality themes and consistent, easy-to-grasp commands should make
working with Cress feel cozy for publishers and maintainers alike.

## Description

Cress is planned to provide:

- Ordinary content pages, landing pages, and list pages.
- Free-form content folders with clean, portable URLs.
- Explicit main and footer navigation.
- Two built-in themes and support for local third-party themes.
- Site and theme initialization, static builds, and a local preview with
  automatic rebuilds.
- JSON schemas and references for configuration and frontmatter.

Cress generates plain static files offline, ready for the hosting and
deployment tools you already use. The
[product requirements](docs/PRD.md#scope) describe the full scope.

## Getting started

Cress is in early development. Requirements, architecture, and planning
are complete, and implementation follows the [work plan](TODO.md).
There is no usable release yet.

To build it from source, follow the [developer guide](docs/DEVELOPER.md).

## Documentation

- [Product requirements](docs/PRD.md): scope, principles, and acceptance
  criteria for 1.0.
- [Architecture guide](docs/ARCHITECTURE.md): system boundary,
  responsibilities, build and preview flow, and failure boundaries.
- [Architectural decision records](docs/adr/): accepted decisions and their
  rationale, listed in the [decision index](docs/PRD.md#decision-records).
- [Developer guide](docs/DEVELOPER.md): set up, build, and test the source.
- [Work plan](TODO.md): implementation phases.

## Contributing

Ideas, questions, and bug reports are welcome as
[issues](https://github.com/cozybadgerde/cress/issues). For code changes,
please open an issue first so we can agree on the approach before a pull
request. The [developer guide](docs/DEVELOPER.md) covers setup and checks.

## Contact

For questions or feedback, contact the project maintainers via:

- Website: [cozybadger.de](https://www.cozybadger.de)
- Email: [devel@cozybadger.de](mailto:devel@cozybadger.de)
- GitHub: [cozybadgerde](https://github.com/cozybadgerde)

## License

Cress is licensed under [BSD-3-Clause](LICENSE). No third-party code or
assets are bundled yet. See
[the licensing decision](docs/adr/0011-bsd-3-clause-license.md).

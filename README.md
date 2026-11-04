# Cress

![Cress: a mouse sheltering beneath a green leaf beside the wordmark.](docs/assets/logo/logo.webp)

A cozy static site generator. Fresh little sites, fast.

Cress turns Markdown content, one TOML configuration file, and a curated
theme into static website files. It is built by
[Cozy Badger](https://www.cozybadger.de) for small businesses and
solopreneurs first.

## Getting started

[Install Cress](docs/INSTALL.md), then create and preview a site:

```sh
cress site init my-site
cress site preview --port 0 my-site
```

Open the printed URL. Edit `my-site/content/` and `my-site/cress.toml`,
then refresh the browser. Stop preview with Ctrl-C.
Run `cress site build my-site` to produce `my-site/public/`.
Deploy those files with your hosting tools.

The [user guide](docs/USER.md) walks through your first page, configuration,
preview, and build.

## What Cress provides

- Ordinary pages, landing pages, and journal-style lists.
- Markdown content with page-owned images and shared static files.
- Three offline themes for freelancers, small businesses, and makers.
- Local theme packages, explicit navigation, and automatic preview rebuilds.

The [product requirements](docs/PRD.md#scope) define the scope.
Release readiness is tracked in the [work plan](TODO.md).

## Documentation

- [User guide](docs/USER.md): install, write, configure, preview, and build.
- [Installation](docs/INSTALL.md): obtain and verify an executable.
- [Theme development](docs/THEMES.md): create a package and use prepared data.
- [Developer guide](docs/DEVELOPER.md): set up, build, and test the source.
- [Product requirements](docs/PRD.md): scope and acceptance criteria.
- [Architecture guide](docs/ARCHITECTURE.md): system boundary,
  responsibilities, build and preview flow, and failure boundaries.
- [Architectural decision records](docs/adr/): accepted decisions and their
  rationale, listed in the [decision index](docs/PRD.md#decision-records).
- [Accessibility assessment](docs/ACCESSIBILITY.md): evidence and blockers.
- [Release maintenance](docs/RELEASE.md): inventory, packaging, and signing.
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

Cress's code is licensed under [BSD-3-Clause](LICENSE). The
[project artwork](docs/assets/logo/README.md) is copyright Daniel Schier
and licensed under CC BY-SA 4.0 International. Bundled dependencies retain
their own licenses; see the
[third-party notices](docs/THIRD-PARTY-NOTICES.md) and
[licensing decision](docs/adr/0011-bsd-3-clause-license.md).

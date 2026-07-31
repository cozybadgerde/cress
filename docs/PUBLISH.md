# Publishing guide

This guide covers getting a built site onto a host. It assumes Cress is
installed (the [operator guide](./OPERATOR.md)) and that you have a site that
builds (the [user guide](./USER.md)).

## What you are publishing

`cress build` renders the site into `public/`. That directory is a complete
static site: HTML, CSS, images, and a `404.html` at its root. There is no
server-side part, nothing to run, and no build step left for the host to do.
Anything that serves files over HTTP will serve it.

There is one thing to get right, and the rest of this guide is mostly about it:
whether your site lives at the root of its domain or in a subdirectory.

## Where the site lives

Cress writes root-relative links, rooted under the path in `base_url`. Only the
path component of that key is load-bearing, which is why it is needed less often
than it looks:

| Where you publish | Example | `base_url` |
| --- | --- | --- |
| User or organization site | `you.github.io` | optional |
| Custom domain | `example.com` | optional |
| Your own server, at the domain root | `example.com` | optional |
| **Project site** | `you.github.io/my-site/` | **required** |

The guidance is not "always set `base_url`". It is **set it when your site does
not live at the root of its domain.**

Making the key mandatory was considered and rejected. Root-relative links mean a
single build serves correctly from every domain pointing at it, so a site
reachable at `example.com`, `example.de`, and `example.org` has no one canonical
host to name. Requiring the key would make the common case pay for the uncommon
one.

A subdirectory is the uncommon one, and it is the default on both GitHub Pages
and GitLab Pages for a project. Get it wrong and the pages still load, but every
link in them points one level too high: the site renders unstyled, with dead
navigation and no images. Nothing in the build can detect this, because where a
site is published is not something the files know.

The [user guide](./USER.md#publishing-to-a-subdirectory) covers the config side.

## GitHub Pages

**First, change the source.** In the repository, go to Settings, then Pages, and
set Source to **GitHub Actions**. The default is a branch, which serves the
repository's files instead of building anything, so the workflow below will run
and publish nothing until this is changed.

Then add `.github/workflows/publish.yml`:

```yaml
name: Publish

on:
  push:
    branches: [main]

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: pages
  cancel-in-progress: false

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7

      - name: Install cress
        run: |
          curl -sSfL https://raw.githubusercontent.com/cozybadgerde/cress/trunk/scripts/install.sh \
            | sh -s -- -b /usr/local/bin -t v1.0.0

      - run: cress build

      - uses: actions/upload-pages-artifact@v3
        with:
          path: public

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - id: deployment
        uses: actions/deploy-pages@v4
```

Notes on the parts that are not obvious:

- **`pages: write` and `id-token: write`** are both required. `deploy-pages`
  authenticates with an OIDC token, so without the second one the deploy fails
  even though the build succeeded.
- **Pin the cress version.** An unpinned install means a release published
  between two runs changes what the second one builds, with no commit to point
  at. See [Pin a version](./OPERATOR.md#pin-a-version).
- **`path: public`** matches cress's default output directory. If you build with
  `--output`, change it here to match.
- **No `.nojekyll` is needed.** That file exists to stop Jekyll from eating
  paths beginning with an underscore, and a GitHub Actions deployment does not
  run Jekyll at all. It is only needed when publishing from a branch source.

For a **project site** at `https://you.github.io/my-site/`, set `base_url`
accordingly:

```toml
[site]
base_url = "https://you.github.io/my-site"
```

For a **user or organization site** from a `you.github.io` repository, or for a
custom domain, leave `base_url` at the domain root or leave it out.

## GitLab Pages

Add `.gitlab-ci.yml`:

```yaml
pages:
  stage: deploy
  image: alpine:latest
  script:
    - apk add --no-cache curl tar
    - |
      curl -sSfL https://raw.githubusercontent.com/cozybadgerde/cress/trunk/scripts/install.sh \
        | sh -s -- -b /usr/local/bin -t v1.0.0
    - cress build
  artifacts:
    paths:
      - public
  rules:
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
```

Notes:

- **`public` is GitLab's default publish directory**, and it is also cress's, so
  there is nothing to move and nothing to configure. The `artifacts: paths` entry
  is what hands it to Pages.
- **The job is named `pages`**, which is the long-standing way to mark a Pages
  deployment. Since GitLab 17.5 any job name works if you add `pages: true` to
  it instead.
- **The image needs `curl`, `tar`, and a SHA-256 tool** for the install script,
  which is why `apk add` is there. Any image providing those works; see the
  [operator guide](./OPERATOR.md#prerequisites).

For a **project site** at `https://namespace.gitlab.io/my-site/`, which is the
default, set the path:

```toml
[site]
base_url = "https://namespace.gitlab.io/my-site"
```

For a **user or group site** from a `namespace.gitlab.io` repository, or for a
custom domain, leave `base_url` at the domain root or leave it out.

## Your own server

Build, then copy the output into the document root:

```bash
cress build --source my-site
rsync -a --delete my-site/public/ user@host:/var/www/my-site/
```

The trailing slash on `my-site/public/` is what makes `rsync` copy the contents
rather than the directory itself. `--delete` removes files on the server that
are no longer in the build, which is what keeps a renamed page from being served
under both names.

A site at the root of its domain needs no `base_url` at all.

## The 404 page

Every build writes a `404.html` at the root of `public/`, carrying the site's
own navigation and styling. Both platforms look for it there, so a project site
gets its own 404 without configuration:

- **GitHub Pages** serves a `404.html` found at the root of what you publish,
  which for the workflow above is the uploaded `public/`.
- **GitLab Pages** tries `/my-site/404.html` first for a project site and falls
  back to `/404.html`, so the file cress writes is the one that answers.

Other hosts mostly do the same, but some need the file named in their own
configuration; check your host's documentation. The
[user guide](./USER.md#the-404-page) covers where that file comes from and how
to replace its wording with your own.

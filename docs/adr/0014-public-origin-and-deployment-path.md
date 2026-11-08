# Separate public URL identity from deployment paths

Status: accepted.

Date: 2026-10-15.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Static sites can be served at a domain root or beneath a path such as
`/mysite/`. GitHub Pages illustrates both cases: organization sites
normally use the account's domain root, while project sites normally use
a repository path. Custom domains can change the publication location.

Navigation and assets must work at the chosen location. Canonical URLs
and sitemap entries also need a public identity. Social metadata such as
Open Graph has similar needs when supported.

These concerns are related but different. A site can be built and browsed
before its public hostname is known, and serving the same files on another
hostname should not inherently break ordinary internal links.

## Decision drivers

- Support domain-root and sub-path hosting without provider-specific rules.
- Keep ordinary browsing independent of the public hostname.
- Produce accurate absolute publication URLs when the origin is known.
- Allow useful generation before a public origin is configured.
- Make missing publication information visible rather than invent it.
- Keep URL resolution in Cress rather than duplicate it in themes.

## Considered options

- Separate an optional public origin from an explicit deployment path.
- Require one full public site URL and derive its origin and path.
- Use document-relative links without configuring a deployment path.

## Decision outcome

Chosen option: "Separate an optional public origin from an explicit
deployment path", because publication identity and browsing paths have
different requirements.

A required full site URL would unnecessarily require a hostname for local
work. Document-relative links can improve portability, but do not replace
an explicit publication path for resolving site-root references and
constructing public metadata.

### Public origin and deployment path

The public origin consists of the scheme, host, and optional port.
The deployment path is a separate site-wide setting.

- Users SHOULD configure a public origin, such as `https://example.org`,
  for canonical URLs, sitemap entries, and other origin-dependent metadata.
- The deployment path defaults to `/`. Users MUST explicitly configure a
  non-root path when publishing beneath one.
- Cress does not infer either value from a hosting provider, organization,
  or repository name.
- The public origin describes publication identity, not an allowed serving
  hostname. Generator-owned internal browsing links do not require it.

An illustrative page with the route `/about/`:

| Public origin | Deployment path | Internal browsing URL | Absolute publication URL |
|---------------|-----------------|-----------------------|--------------------------|
| `https://example.org` | `/` | `/about/` | `https://example.org/about/` |
| `https://example.org` | `/mysite/` | `/mysite/about/` | `https://example.org/mysite/about/` |
| Not configured | `/mysite/` | `/mysite/about/` | Omitted with a warning |

The browsing examples show the resolved destinations, not a requirement
to emit root-relative rather than document-relative HTML links.

### Missing public origin

Without a public origin, generate a usable site and issue an actionable
warning. Omit metadata that requires that origin rather than invent a
hostname or emit incomplete publication URLs.

Ordinary content, titles, descriptions, and internal browsing remain
available. Do not advertise an omitted sitemap through generated
`robots.txt`. Values that already provide valid absolute URLs do not
require an invented origin.

This does not add Open Graph or other social metadata fields to 1.0.
When supported, their origin-dependent URLs follow the same policy.

### URL resolution responsibilities

- Treat site-root Markdown links and images as relative to the site's
  deployment root. Under `/mysite/`, `/about/` resolves to
  `/mysite/about/`, not the host's `/about/`.
- Resolve source-relative Markdown links and images to their published
  destinations, as established in ADR 0003.
- Apply the deployment path consistently to generator-owned page,
  navigation, and asset URLs. Combine the public origin, deployment path,
  and published route for absolute publication URLs.
- Leave external URLs unchanged.
- Keep destination and prefix resolution in Cress. Themes use resolved
  URLs rather than implement their own deployment rules.

Copied asset contents remain unchanged. Authors must make references
inside CSS and other copied assets suitable for their published location;
Cress does not rewrite those files to repair hostname or path assumptions.

### Authored HTML

Cress resolves Markdown links and images only. URLs in authored HTML stay
as written, as the [raw HTML decision](0015-trusted-markdown-and-raw-html.md)
defines.

For example, with a deployment path of `/mysite/`:

| Authored reference | Handling |
|--------------------|----------|
| `[About](/about/)` | Resolve to `/mysite/about/`. |
| `![Product](/images/product.jpg)` | Resolve to `/mysite/images/product.jpg`. |
| `<a href="/about/">About</a>` | Leave `href="/about/"` unchanged. |
| `<img src="/images/product.jpg" alt="Product">` | Leave `src="/images/product.jpg"` unchanged. |
| `[External](https://example.net/)` | Leave the external destination unchanged. |

Output built for one deployment path does not have to work unchanged at
another.

### Consequences

- Good, because sites can be built before a public hostname is known.
- Good, because browsing does not depend on that hostname.
- Good, because deployment paths and publication identity are explicit.
- Good, because themes share one URL-resolution policy.
- Bad, because missing origins leave publication metadata unavailable.
- Bad, because an incorrectly configured path can break browsing.
- Bad, because copied assets must not assume an unsuitable hosting path.
- Bad, because authored HTML links can break without link-check warnings.

## Confirmation

During implementation, check generation at `/` and a non-root path.
Confirm page links, navigation, and assets work on different serving
hostnames without changing the configured public identity.

Check absolute publication URLs include the configured path exactly once.
Without an origin, confirm successful generation, actionable warnings,
omission of origin-dependent metadata, and no dangling sitemap reference.
Check site-root and source-relative Markdown links and images, unchanged
external URLs, unchanged copied asset contents, and unchanged authored
HTML URLs.

## More information

See [content and routes](0003-content-tree-and-routes.md),
[rendering responsibilities](0008-rendering-and-template-trust-boundary.md),
and [navigation](0013-explicit-site-navigation.md).

Hosting examples follow
[GitHub Pages site types](https://docs.github.com/en/pages/getting-started-with-github-pages/what-is-github-pages).
[Open Graph](https://ogp.me/) defines `og:url` as the object's canonical
identity, which link previews use.

# Accessibility assessment

## Scope and release status

**Release blocked: criterion-by-criterion assessment is pending.** Cress targets
WCAG 2.2 Level AA for complete default pages, not arbitrary customized sites.
No conformance or legal-compliance claim is made.

The assessment fixtures are the seven English pages created by `site init`
with each of `poppy`, `clover`, and `anemone`: home landing, services with a table,
about, contact with opening hours, journal list, and dated article with
captioned code, plus a content tour with Markdown and authored HTML examples.
The tour also demonstrates automatic and explicit Markdown heading targets
with same-page links. Verify fragment navigation and heading reading order
without treating IDs as additional interactive controls.
The tour includes strikethrough and static task lists with labeled native
disabled checkboxes. Check announced names, completion state, disabled
state, nested-list reading order, and usable links inside task text.
These are not interactive forms. Confirm important changes remain clear
when strikethrough is not announced.
The home includes an original captioned SVG. Automated
tests additionally cover German labels, supplied metadata, title images,
caption links, and page-type composition. No scripts, forms, media players,
authentication, menus that open/close, or timed interactions are shipped.

Automated evidence recorded on 2026-10-21: Go unit and filesystem tests on
macOS pass for default palette contrast, one main landmark and page heading,
skip target, labeled navigation, local assets, page types, and initialization.
`mise run check` runs those tests. Contrast checks use WCAG relative
luminance: at least 4.5:1 for text, including syntax colors, and 3:1 for
boundaries. These checks do not measure rendered pixels or browser behavior.

On 2026-10-30, Anemone passed the same automated palette, markup, localization,
and initialized-site checks. Filesystem tests also cover image-free pages,
an empty list, German dates, and a deployment prefix.

A scoped visual review on macOS with Firefox 158.0.1 inspected Anemone at
1280 and 320 CSS pixels: default landing and list pages, the content tour,
image-free and title-image landings, long titles, links, code, and tables.
Desktop inspection confirmed image-only rotation, upright captions, and
the side masthead. Narrow screenshots confirmed stacked layout, unrotated
images, wrapped long content, and table overflow contained in its wrapper.
These one-off checks are not a complete-page accessibility assessment.

On 2026-11-01, the maintainer reports manual exploratory accessibility testing
across the three themes with a "works for me" result. No failures were reported.
The report does not specify the tested pages, executable version, browser,
assistive technology, keyboard, touch, zoom, or criterion-specific outcomes.
It records exploratory evidence, not completion of the A/AA matrix below.
A persistent browser automation suite remains outside the agreed scope.

On 2026-11-04, a scoped Firefox 159.0 headless review on macOS compared
short/long identities, English/German navigation labels, and title-, body-,
inline-, code-, and list-image captions in all three themes. Rendered
measurements at 320, 640, 768, 960, and 1440 CSS pixels found no horizontal
page overflow, including text-spacing overrides at 320px. Identity and
navigation targets measured at least 44px in each dimension; muted caption
text exceeded 4.5:1 against the default backgrounds. Screenshots also covered
390px layouts. Tab input followed the skip link, identity, and main links in
document order. The headless session did not expose `:focus-visible` states,
so visible keyboard-focus verification remains part of the human review.
These exploratory checks do not complete the release matrix.

## Human review procedure

Generate fresh fixtures using the current executable:

```sh
cress site init --theme poppy poppy-review
cress site init --theme clover clover-review
cress site init --theme anemone anemone-review
cress site preview --port 0 poppy-review
```

Stop each preview before reviewing the next theme with the equivalent command.
Use each printed URL. Also review German exact-language labels by changing
`site.language` to `de`; sample page content itself remains English and
must be translated or language-marked for a German-language assessment.

For each theme and page, record executable version/commit, OS, browser and
version, assistive technology/version, viewport and zoom settings, reviewer,
date, observations, applicability, failures, and remediation/retest results.
Do not mark a criterion passed from source inspection alone.

Check keyboard-only navigation, skip-link focus, link purpose, focus
visibility and order, screen-reader landmarks/headings/tables/images,
200% text resizing, 400% zoom/reflow at 320 CSS pixels, text spacing
overrides, landscape/portrait, target sizes, contrast in rendered states,
and user color/forced-color modes where supported. Verify long titles,
navigation labels, captions, code lines, and table contents. Review raw
HTML and customizations separately when they are present.

## WCAG 2.2 A/AA review matrix

Every row remains pending human confirmation. "Not present" is a proposed
applicability decision for these fixtures, not a universal exemption.
Record a separate outcome for all three themes and every affected page.

| Criteria | Current evidence or applicability to confirm |
|----------|---------------------------------------------|
| 1.1.1 Non-text content | Sample SVG has descriptive alternative; inspect equivalent purpose and caption. |
| 1.2.1–1.2.5 Audio/video | No audio/video in fixtures; confirm not applicable. |
| 1.3.1 Info and relationships | Markup checks cover landmarks/headings; inspect table headers and accessibility tree. |
| 1.3.2 Meaningful sequence | Inspect reading order for all page types. |
| 1.3.3 Sensory characteristics | Review content instructions. |
| 1.3.4 Orientation | Check portrait and landscape without restriction. |
| 1.3.5 Identify input purpose | No personal-data entry; task checkboxes are disabled. Confirm applicability. |
| 1.4.1 Use of color | Links are underlined; inspect all states and content. |
| 1.4.2 Audio control | No audio; confirm not applicable. |
| 1.4.3 Contrast (minimum) | Default CSS text palette passes automated ratios; inspect rendered states. |
| 1.4.4 Resize text | Check 200% text resizing. |
| 1.4.5 Images of text | SVG contains no text; inspect all fixtures. |
| 1.4.10 Reflow | Check 320 CSS pixel width and 400% zoom. |
| 1.4.11 Non-text contrast | CSS boundaries pass ratios; inspect focus indicators and imagery. |
| 1.4.12 Text spacing | Apply spacing overrides and check clipping/overlap. |
| 1.4.13 Content on hover or focus | No extra popovers/tooltips; confirm not applicable. |
| 2.1.1 Keyboard | Exercise every link and scrollable area. |
| 2.1.2 No keyboard trap | Verify entry/exit and skip-link behavior. |
| 2.1.4 Character key shortcuts | No shortcuts; confirm not applicable. |
| 2.2.1 Timing adjustable | No time limits; confirm not applicable. |
| 2.2.2 Pause, stop, hide | No moving/auto-updating page content; confirm not applicable. |
| 2.3.1 Three flashes or below threshold | No flashing; confirm not applicable. |
| 2.4.1 Bypass blocks | Skip link and one main target are tested; verify actual focus movement. |
| 2.4.2 Page titled | Templates emit escaped page/site titles; inspect browser titles. |
| 2.4.3 Focus order | Verify logical order at every layout. |
| 2.4.4 Link purpose (in context) | Review sample and navigation link text. |
| 2.4.5 Multiple ways | Main/footer navigation and ordinary links connect the sample pages; verify findability. |
| 2.4.6 Headings and labels | Review descriptive headings and navigation labels. |
| 2.4.7 Focus visible | CSS defines focus outlines; inspect keyboard states. |
| 2.4.11 Focus not obscured (minimum) | No sticky headers; verify focused links remain visible. |
| 2.5.1 Pointer gestures | No multipoint/path gestures; confirm not applicable. |
| 2.5.2 Pointer cancellation | Standard links; check normal activation/cancellation. |
| 2.5.3 Label in name | Check accessible link and checkbox names match visible labels. |
| 2.5.4 Motion actuation | No motion controls; confirm not applicable. |
| 2.5.7 Dragging movements | No dragging; confirm not applicable. |
| 2.5.8 Target size (minimum) | Navigation has padded targets; measure rendered targets and inline exceptions. |
| 3.1.1 Language of page | Configured `lang` is tested; verify content language. |
| 3.1.2 Language of parts | Review author content and code applicability. |
| 3.2.1 On focus | No scripted navigation; verify no unexpected changes. |
| 3.2.2 On input | Task checkboxes are disabled; confirm no input-driven changes. |
| 3.2.3 Consistent navigation | Shared templates preserve order; inspect across page types. |
| 3.2.4 Consistent identification | Review repeated links and identity. |
| 3.2.6 Consistent help | Contact link repeats in footer; inspect position and ordering. |
| 3.3.1–3.3.4 Input errors/instructions/prevention | No forms; confirm not applicable. |
| 3.3.7 Redundant entry | No data entry; confirm not applicable. |
| 3.3.8 Accessible authentication (minimum) | No authentication; confirm not applicable. |
| 4.1.2 Name, role, value | Inspect links, landmarks, and checkbox names, completion, and disabled states. |
| 4.1.3 Status messages | No dynamic page status messages; terminal preview diagnostics are not injected. |

WCAG 2.2 removes 4.1.1 Parsing; it is not an additional 2.2 criterion.
Known applicable A/AA failures block release until fixed and retested.

## Publisher responsibilities

Use meaningful content headings below the theme's page-level heading.
Supply useful image alternatives or an explicit empty alternative for
decorative images. Captions are not replacements for alternatives.
Review translations and language changes within content.
Give every task meaningful text. Explain important changes in words
rather than relying on strikethrough alone.

Built-in themes use the brand color as their accent for links, buttons,
and Clover's title block. Cress warns when it has less than 4.5:1 contrast
against white as a reference only, not an assessment of the selected theme.
Dark or tinted themes may intentionally use colors with low contrast
against white. Check actual text/background pairs whether or not a warning
appears; Poppy's off-white and Anemone's paper backgrounds also change
the measured contrast. In Anemone, inspect framed images at wide and narrow
widths, unrotated captions, image-free pages, and side-masthead reading order.
Arbitrary custom CSS, raw HTML,
third-party themes, changed navigation/content, or brand assets need
assessment of the resulting complete pages. Passing tests is not
certification of a finished site.

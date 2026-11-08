# Accessibility guide

A website is accessible when people can use it however they browse. Some
readers are blind and listen to a screen reader. Others navigate with a
keyboard or switch device, zoom text to 200 percent, read on a small phone in
bright sunlight, or need plain language and a predictable layout. Many people
fit several of these at once, and most of us fit one of them at some point,
with a broken arm, a tired evening, or an old laptop.

Cress tries to make the accessible result the default one. It follows common
best practices and uses the
[Web Content Accessibility Guidelines (WCAG) 2.2](https://www.w3.org/TR/WCAG22/)
at level AA as a reference. The [product requirements](PRD.md#accessibility)
describe how responsibility is shared.

## Meaningful structure

Screen readers let people jump between headings, landmarks, and links, much
like sighted readers skim a page. That only works when the markup says what
each part is.

- Every page declares its language with `lang`, so screen readers pronounce
  it correctly.
- The bundled themes use `header`, `nav`, `main`, and `footer` landmarks.
  Navigation areas have names such as "Main navigation".
- The current page in the navigation carries `aria-current="page"`.
- Markdown headings become real headings with stable anchors.
- Captioned images become `figure` and `figcaption`. Task lists use native,
  labelled checkboxes, and strikethrough uses `del`.
- Decorative separators are hidden from screen readers.

## Keyboard use

Many people never touch a mouse. Everything that works with a pointer should
work with <kbd>Tab</kbd>, <kbd>Enter</kbd>, and the arrow keys.

- A skip link, visible on focus, jumps straight to the main content.
- Focus outlines are clearly visible on links and other focusable areas.
- Wide tables can receive focus, so keyboard users can scroll them.
- The bundled themes ship no JavaScript, so there are no custom widgets that
  trap focus or change the page unexpectedly.

## Readable presentation

Text must stand out from its background, and meaning must not depend on
color alone.

- The default theme palettes are tested for a contrast of at least 4.5:1.
- Links are underlined and the current page is bold, so neither relies on
  color alone.
- Cress warns when a configured brand color has low contrast against white.
  Check the result in the actual theme as well.
- In Anemone, Windows high-contrast mode keeps image borders visible, and
  reduced-motion settings turn off image effects.

## Adaptable layout

People zoom, enlarge fonts, rotate phones, and adjust text spacing. Pages
should reflow rather than break.

- Layouts use relative units and adapt down to phone widths without
  sideways page scrolling.
- Long words and URLs wrap instead of overflowing.
- Wide tables and code blocks scroll within their own area.
- Navigation links are at least 44 by 44 CSS pixels, which helps on touch
  screens and for people with limited fine motor control.

## Understandable and predictable

Clear wording and a consistent layout help everyone, especially people with
cognitive disabilities or those reading in a second language.

- Navigation stays in the same place on every page and is set explicitly by
  the publisher.
- Theme labels such as "Skip to content" come in English and German.
- Dates use the `time` element with a machine-readable value.

## What publishers can do

Cress can only preserve what the content provides.

- Use one main heading per page and keep heading levels in order.
- Write link text that makes sense on its own, rather than "click here".
- Give every informative image alternative text that conveys its purpose.
  Use empty alternative text for purely decorative images. A caption does
  not replace the alternative.
- Set the site language, and review translated labels and content.
- Explain important changes in words, not only with strikethrough.
- Check contrast after changing brand colors or CSS.
- Review the result when adding raw HTML or custom styles.

## What theme authors can do

Third-party themes own their markup and styles. The bundled themes are good
starting points.

- Keep landmarks, the skip link, and visible focus styles.
- Respect user settings such as `prefers-reduced-motion` and
  `forced-colors`.
- Test with a keyboard, at 200 percent zoom, and on a narrow screen.
- See the [theme developer guide](THEMES.md) for the template contract.

## Learn more

- [WCAG 2.2 at a glance](https://www.w3.org/WAI/standards-guidelines/wcag/glance/)
- [Easy checks: a first review of web accessibility](https://www.w3.org/WAI/test-evaluate/easy-checks/)
- [WebAIM contrast checker](https://webaim.org/resources/contrastchecker/)

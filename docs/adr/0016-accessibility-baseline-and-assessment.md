# Set an accessibility baseline for generated default pages

Status: accepted.

Date: 2026-10-15.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Cress produces markup that named themes combine with publisher content.
The project needs an actionable accessibility release expectation without
treating the generator, a theme, or a successful build as proof that an
arbitrary finished site is accessible.

The PRD already targets Web Content Accessibility Guidelines (WCAG) 2.2
Level AA. That target needs an assessment scope, a response to known
failures, and a boundary between release obligations and later improvements.
Accessibility also matters to publishers with legal obligations, but a
technical accessibility assessment cannot establish all legal compliance.

## Decision drivers

- Make the existing accessibility target actionable before implementation.
- Assess complete pages rather than claim conformance for isolated parts.
- Preserve generator, theme, and publisher ownership.
- Combine repeatable checks with knowledgeable human evaluation.
- Avoid overstating assessment coverage or promising automatic compliance.
- Distinguish baseline requirements from additional improvements.

## Considered options

- Assess default pages against WCAG 2.2 A and AA with a release gate.
- Keep a best-effort target and permit known failures with documentation.
- Rely on automated accessibility checks alone.

## Decision outcome

Chosen option: "Assess default pages against WCAG 2.2 A and AA with a
release gate", because it gives the existing target a concrete release
expectation without extending assurance to arbitrary publisher sites.

### Assessment scope and release expectation

Assess complete default pages produced with each theme shipped with Cress,
covering supported page types and their default content and interactions.
Use the combined output of generator, theme, and example content as the
assessment unit. Component checks contribute evidence but do not establish
complete-page conformance.

The baseline includes all applicable WCAG 2.2 Level A and Level AA success
criteria, not only a selected list of easy checks.
Known applicable A or AA failures in the assessed default pages block the
release until fixed. Do not defer them while claiming AA.

Apply this expectation before 1.0 and to subsequent affected releases.
Document assessment scope, environments, results, applicability decisions,
and limitations. A passing build or automated scan is not a conformance
certificate. Any conformance claim must meet WCAG's conformance
requirements and remain scoped to the assessed pages and conditions.

### Ownership and limits

Apply the [existing ownership boundary](0008-rendering-and-template-trust-boundary.md):
Cress owns markup it generates directly and preservation of author-supplied
accessibility information; each theme owns its presentation; publishers
own their content and customizations.

Bundled and third-party themes have the same presentation responsibilities.
The project's release assessment covers the themes and default pages it
ships, not certification of third-party packages or arbitrary sites.
Publisher changes to content, colors, HTML, styles, or themes require
review of the resulting pages.

Keep [authored HTML trusted and unchanged](0015-trusted-markdown-and-raw-html.md).
This baseline does not add sanitization, automatic repair, or a generator
accessibility-validation command. Final-site legal applicability remains
the publisher's responsibility; WCAG assessment is not a blanket guarantee
of compliance with every applicable law.

[Caption, attribution, and disclosure support](../prd.md#captions-attribution-and-disclosures)
is defined in the PRD. Captions remain distinct from image alternatives;
publishers determine applicable disclosure obligations and wording.

### Assessment approach

Run linters and automated tests in continuous integration (CI).
Use unit tests as the default. Add integration tests where exercising real
boundaries adds meaningful confidence. Automated end-to-end tests and
browser automation are outside the initial scope.

Complement automation with documented human accessibility review.
Establish the baseline before 1.0 and review relevant changes to generated
markup or theme behavior. Linters and automated tests alone do not
establish WCAG conformance. Concrete tools and check procedures belong to
implementation planning.

Assessment is development and release work, not an additional validation
workflow imposed on people generating their own sites.

### Later improvements

Additional Level AAA goals, broader compatibility coverage, and usability
studies can follow. They do not replace the baseline or justify deferring
known applicable A or AA failures.

### Consequences

- Good, because the release target has a defined scope and failure policy.
- Good, because automated and manual review address different limitations.
- Good, because claims distinguish assessed output from arbitrary sites.
- Bad, because assessment requires time, expertise, and recorded evidence.
- Bad, because changes can require renewed manual review.
- Neutral, because publishers remain responsible for their finished sites.

## Confirmation

Before releasing, confirm the assessment covers the shipped themes and
default page types, records automated and manual results, and leaves no
known applicable A or AA failures in the assessed output.

Check that claims identify their scope and limitations, relevant changes
receive renewed assessment, and documentation explains publisher and
third-party theme responsibilities. These are future release obligations,
not evidence that an implementation or assessment already exists.

## More information

See [the accessibility product requirements](../prd.md#accessibility),
[theme integration](0007-theme-packages-and-integration.md),
and [release distribution](0012-release-artifacts.md).

[Understanding WCAG conformance](https://www.w3.org/WAI/WCAG22/Understanding/conformance)
explains levels and complete-page requirements.
[Evaluating web accessibility](https://www.w3.org/WAI/test-evaluate/)
explains why automated checks need human evaluation.

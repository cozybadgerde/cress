# License the project under BSD-3-Clause

Status: accepted.

Date: 2026-10-06.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Cozy Badger wants Cress to be an open-source product usable for commercial
small-business sites and third-party theme development.
The owner prefers BSD-3-Clause or a similar permissive license.

## Decision drivers

- Permit commercial use, modification, and redistribution.
- Keep license obligations understandable.
- Retain attribution and non-endorsement conditions.
- Separate project licensing from third-party asset rights.
- Keep dependency license approvals explicit and auditable.
- Publish a machine-readable inventory of bundled components and licenses.

## Considered options

- BSD-3-Clause.
- MIT.
- Apache-2.0.

## Decision outcome

Chosen option: "BSD-3-Clause", because it matches the owner's preference for
a permissive license with attribution and non-endorsement conditions.

Apply BSD-3-Clause to project-owned code and documentation, including
project-owned built-in theme code.
Include the full license text and correct copyright notices before release.
This decision records intent; it does not itself add a license file.

Review dependencies, fonts, photographs, and other bundled assets under
their own licenses. Third-party themes may have different licenses.
The generator license does not determine the license of user-authored
content or all generated output.

### Dependency license policy

The initial approved licenses for bundled code dependencies are:

- `MIT`.
- `BSD-2-Clause`.
- `BSD-3-Clause`.
- `Apache-2.0`.

Use SPDX license identifiers. Approval permits inclusion under the
dependency's own license; it does not relicense third-party code as
BSD-3-Clause or remove notice, attribution, or other applicable obligations.
Review direct and transitive dependencies against their actual license
terms, including license expressions and exceptions.

Additional licenses require explicit review and approval before inclusion.
Unknown or unapproved licenses cannot pass release checks automatically.
Record approved additions and their rationale in the licensing policy.

This code allowlist does not automatically approve fonts, photographs,
or other non-code assets. Review their distribution rights separately.
User-provided content and third-party themes not bundled with Cress are
outside the release inventory.

### Machine-readable license inventory

Maintain a software bill of materials (SBOM) for bundled components,
including transitive code dependencies and embedded theme assets.
Use a standard machine-readable format. Select concrete tooling and
representation with the release implementation.

Refresh the inventory when bundled components change and generate
release-specific SBOM artifacts from the actual release inputs.
Keep bundled components traceable to their sources, licenses, and review
outcomes.
Do not treat automated license detection as proof of approval.

Publish the SBOM artifacts with each release and include them in the
authenticated checksum manifest. Release checks reject unreviewed
components and unapproved licenses. Ship required license texts and notices
separately; an SBOM does not replace them.

BSD-3-Clause does not contain the explicit patent grant found in
Apache-2.0. This trade-off is acknowledged rather than treating the
licenses as interchangeable.
This decision does not grant use of Cozy Badger trademarks.

### Consequences

- Good, because commercial and third-party use fit the product goal.
- Good, because the project uses a familiar permissive license.
- Bad, because bundled assets still need a separate rights review.
- Bad, because inventory and approval records must track dependency changes.
- Neutral, because an explicit patent grant would require reconsideration.

## Confirmation

Before release, confirm that the license file and notices match this
decision. Review distributable dependencies and bundled assets for their
own obligations. Document any assets excluded from the project license.
Verify direct and transitive bundled components appear in the release SBOM
with licenses and review records. Verify release checks reject unreviewed
components and unapproved licenses, and that published SBOM artifacts are
covered by the authenticated checksum manifest.

## More information

See the [release requirements](../prd.md) and
[SPDX BSD-3-Clause reference](https://spdx.org/licenses/BSD-3-Clause.html).
See [release artifacts](0012-release-artifacts.md) for SBOM distribution.

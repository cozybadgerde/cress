# Release guide

Follow these steps to publish a Cress release. The maintainer performs every
step that tags, publishes, or deletes something.

## 1. Choose the version

Use a full semantic version with a `v` prefix, such as `v1.2.0`. Add a
prerelease suffix such as `-rc.1` to publish a GitHub prerelease. Build
metadata alone, such as `+build.1`, does not make a prerelease. The
[release workflow](../.github/workflows/release.yml) owns the exact tag rules.

Check that no tag, release, or draft with that version exists yet:

```sh
tag=v1.2.0
git ls-remote --tags origin "$tag"
gh release list --repo cozybadgerde/cress
```

Replace `v1.2.0` with the intended version.

## 2. Go through the pre-tag checklist

Do this for every major and minor release. For patch releases, check what
changed. Note the version, the commit, and anything unusual in the release
issue or pull request.

Use the intended release revision and fresh test sites from
[the developer guide](DEVELOPER.md#test-sites). For the accessibility rows,
try a landing page and the content tour in each theme. The
[accessibility guide](ACCESSIBILITY.md) explains what to look for.

| # | Check | What to look for | Notes |
|---|-------|------------------|-------|
| 1 | Source is ready | Local `trunk` matches remote. No stray fixtures, generated output, or secrets in tracked or leftover files. | |
| 2 | Scope and version are clear | Changes since the last release match the [PRD](PRD.md) and [compatibility policy](adr/0004-shared-contract-policy.md). Breaking changes, migrations, and known limits are described. | |
| 3 | Automated checks are green | [Checks, CI](DEVELOPER.md#testing-and-checks), and build pass for the release commit. Warnings make sense. | |
| 4 | Maintainer workflow works | Following the [user guide](USER.md) and [theme developer guide](THEMES.md#create-your-first-theme): initialize a site and theme, build, preview, edit, refresh, stop. Help and errors are clear. Works offline. | |
| 5 | Content and deployment look right | Pages, lists, images, captions, code, navigation, and English/German labels. Domain-root and URL-prefix output, metadata, sitemap, and robots. | |
| 6 | Safety behavior works | On a test site: an invalid edit, draft preview, and a removed page give useful errors, keep the last good preview, isolate drafts, and remove stale output. | |
| 7 | Keyboard navigation works | Tab through navigation and content links, and try the skip link. Focus is visible, order makes sense, no traps. | |
| 8 | Screen reader makes sense | With VoiceOver or NVDA, headings, navigation, images, a table, and task lists read sensibly. | |
| 9 | Zoom and narrow layouts work | At 200% zoom and phone width, text and links stay readable without sideways page scrolling. | |
| 10 | Tables and code stay usable | Wide content scrolls within its area, also by keyboard. | |
| 11 | Touch works | If a touch device is at hand: navigation and links are easy to use in both orientations. | |
| 12 | Visual clarity is good | Text, links, and focus have readable contrast. Meaning doesn't depend on color alone. | |
| 13 | Problems are handled | Barriers and bugs found above are fixed and retried, or noted as known limits. | |
| 14 | Distribution is ready | Target acceptance, [license reports](#license-reports), and [asset rights](adr/0011-bsd-3-clause-license.md) are fine. The [installation guide](INSTALL.md) is accurate. | |
| 15 | Ready to tag | Open issues are acceptable, and you know how to handle a failed run (step 5). | |

## 3. Rehearse locally

Run from the repository root:

```sh
mise run check
mise run build
mise run check:goreleaser
mise exec -- goreleaser release --snapshot --clean --skip=publish,sign
mise run test:release
```

The snapshot replaces `dist/` and writes `bin/THIRD-PARTY-NOTICES.txt`. It
does not tag, publish, or sign. The smoke test checks archive contents,
binary targets, license reports, and checksums. It runs the host executable
and builds a site with each bundled theme. Other targets are only inspected,
so native-platform checks remain part of checklist row 14.

## 4. Tag and push

```sh
git switch trunk
git pull --ff-only
git tag -a "$tag" -m "Cress $tag"
git push origin "$tag"
```

The push starts the [release workflow](../.github/workflows/release.yml):

1. It validates the tag and runs the checks and build.
2. It stops if a release or draft already exists for the tag.
3. GoReleaser builds the archives and license reports, writes
   `checksums.txt`, signs it with Cosign, and uploads everything to a draft.
4. The archive smoke test runs against the built files.
5. The workflow compares the uploaded digests with the tested files and
   publishes that exact draft. Nothing is rebuilt in between.

Watch the run:

```sh
gh run watch --repo cozybadgerde/cress
```

## 5. If the run fails

A failed run can leave drafts behind. Don't publish them by hand.

1. Read the failed run's log and fix the cause.
2. List all releases and drafts for the tag with their IDs:

   ```sh
   gh api repos/cozybadgerde/cress/releases \
     --jq ".[] | select(.tag_name == \"$tag\") | [.id, .draft, .created_at] | @tsv"
   ```

3. Delete each failed draft by ID. Several drafts can share a tag, so don't
   delete by tag name:

   ```sh
   gh api -X DELETE repos/cozybadgerde/cress/releases/<id>
   ```

4. Rerun the failed workflow. If the fix needs a new commit, release it
   under a new version.

The workflow refuses to start while any release or draft for the tag exists.

## 6. Verify the published release

Download the archive and follow [installation guide](INSTALL.md).
The signature must match this certificate identity and issuer:

```text
https://github.com/cozybadgerde/cress/.github/workflows/release.yml@refs/tags/<exact-tag>
https://token.actions.githubusercontent.com
```

Also confirm that a modified checksum file or a wrong identity fails
verification.

## Reference

### Tools

The workflow and local commands use the tool versions pinned in
[`mise.toml`](../mise.toml). The [release configuration](../.goreleaser.yml)
owns target platforms, archive contents, and report names.

### License reports

GoReleaser runs [go-licenses](https://github.com/google/go-licenses) to
create a `.licenses.csv` report for each platform. Each row holds a library
path, license URL, and license name, without a header. The signed checksum
file covers these reports.

The same tool fills [the notices template](../.github/notices.tpl) to create
the license texts shipped in the archives. `mise run check:licenses`, part
of `mise run check`, fails on unknown or unapproved licenses for all three
targets. The [license policy](adr/0011-bsd-3-clause-license.md#dependency-license-policy)
lists the approved license types. Bundled non-code assets need their own
rights review.

Chroma's combined license file lists OFL-1.1 for a font formatter Cress does
not use. OFL-1.1 is approved, and the notices keep that font's license text.

To print a report locally:

```sh
mise exec -- go-licenses report ./cmd/cress
```

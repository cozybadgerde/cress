#!/usr/bin/env bash
set -euo pipefail

releases() {
  gh api --paginate "repos/$GITHUB_REPOSITORY/releases?per_page=100" |
    jq -sc --arg tag "$RELEASE_TAG" '[.[][] | select(.tag_name == $tag)]'
}

case "${1:?expected check or publish}" in
  check)
    matches=$(releases)
    if [[ $(jq 'length' <<<"$matches") != 0 ]]; then
      echo "A release or draft already exists for $RELEASE_TAG; inspect it before retrying." >&2
      exit 1
    fi
    ;;
  publish)
    matches=$(releases)
    if [[ $(jq 'length' <<<"$matches") != 1 ]]; then
      echo "Expected exactly one draft for $RELEASE_TAG; inspect releases before publishing." >&2
      exit 1
    fi
    id=$(jq -er '.[0].id | select(type == "number")' <<<"$matches")
    release=$(gh api "repos/$GITHUB_REPOSITORY/releases/$id")
    prerelease=false
    if [[ ${RELEASE_TAG%%+*} == *-* ]]; then
      prerelease=true
    fi
    if ! jq -e --arg tag "$RELEASE_TAG" --argjson pre "$prerelease" \
      '.tag_name == $tag and .draft == true and .prerelease == $pre and
       all(.assets[]; .state == "uploaded" and (.digest | type == "string"))' \
      <<<"$release" >/dev/null; then
      echo "Draft $id has an unexpected tag, status, prerelease flag, or asset state." >&2
      exit 1
    fi
    local_assets=$(
      cd "${CRESS_RELEASE_DIR:-dist}"
      sha256sum ./*.tar.gz ./*.licenses.csv ./checksums.txt ./checksums.txt.sigstore.json |
        jq -Rsc '[split("\n")[] | select(length > 0) |
          capture("^(?<hash>[0-9a-f]{64})  \\./(?<name>.+)$") |
          {name, digest: ("sha256:" + .hash)}] | sort_by(.name)'
    )
    remote_assets=$(jq -c '[.assets[] | {name, digest}] | sort_by(.name)' <<<"$release")
    if [[ "$local_assets" != "$remote_assets" ]]; then
      echo "Draft $id assets do not match the tested local artifacts." >&2
      exit 1
    fi
    gh api "repos/$GITHUB_REPOSITORY/git/ref/tags/$RELEASE_TAG" >/dev/null
    result=$(gh api --method PATCH "repos/$GITHUB_REPOSITORY/releases/$id" -F draft=false)
    if ! jq -e --argjson id "$id" --arg tag "$RELEASE_TAG" --argjson pre "$prerelease" \
      '.id == $id and .tag_name == $tag and .draft == false and .prerelease == $pre' \
      <<<"$result" >/dev/null; then
      echo "Publication response for draft $id was unexpected; inspect GitHub." >&2
      exit 1
    fi
    ;;
  *)
    echo "Expected check or publish, got: $1" >&2
    exit 1
    ;;
esac

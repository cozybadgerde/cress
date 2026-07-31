#!/bin/sh
# install.sh — download, verify, and install the cress binary.
# Project: https://github.com/cozybadgerde/cress
#
# Quick install (rootless, into ~/.local/bin):
#   curl -sSfL https://raw.githubusercontent.com/cozybadgerde/cress/trunk/scripts/install.sh \
#     | sh -s -- -b "$HOME/.local/bin"
#
# System-wide (into /usr/local/bin):
#   curl -sSfL https://raw.githubusercontent.com/cozybadgerde/cress/trunk/scripts/install.sh | sudo sh
#
# Pin a version:
#   ... | sh -s -- -t v1.0.0
#
# Installing via curl/wget (as above) avoids the macOS Gatekeeper quarantine
# that a browser download triggers — the binary runs without the "cannot be
# verified" prompt.
#
# Flags / env:
#   -b DIR   install directory   (default /usr/local/bin; or env CRESS_BINDIR)
#   -t TAG   release tag          (default: latest; or env CRESS_VERSION)

set -eu

OWNER="cozybadgerde"
REPO="cress"
BASE="https://github.com/${OWNER}/${REPO}"
API="https://api.github.com/repos/${OWNER}/${REPO}"

BINDIR="${CRESS_BINDIR:-/usr/local/bin}"
TAG="${CRESS_VERSION:-}"

log() { printf '%s\n' "cress-install: $*" >&2; }
die() { log "error: $*"; exit 1; }

usage() {
	cat >&2 <<-EOF
		install cress — a cozy static site generator.

		usage: install.sh [-b DIR] [-t TAG]
		  -b DIR   install directory (default ${BINDIR}; env CRESS_BINDIR)
		  -t TAG   release tag to install (default: latest; env CRESS_VERSION)
	EOF
}

while [ $# -gt 0 ]; do
	case "$1" in
	-b) BINDIR="${2:?-b needs a directory}"; shift 2 ;;
	-t) TAG="${2:?-t needs a tag}"; shift 2 ;;
	-h | --help) usage; exit 0 ;;
	*) die "unknown argument: $1 (try -h)" ;;
	esac
done

# Pick a downloader and the checksum tool, with fallbacks.
if command -v curl >/dev/null 2>&1; then
	dlo() { curl -sSfL -o "$2" "$1"; }
	dls() { curl -sSfL "$1"; }
elif command -v wget >/dev/null 2>&1; then
	dlo() { wget -qO "$2" "$1"; }
	dls() { wget -qO- "$1"; }
else
	die "need curl or wget"
fi
command -v tar >/dev/null 2>&1 || die "need tar"
if command -v sha256sum >/dev/null 2>&1; then
	sum() { sha256sum "$1" | awk '{print $1}'; }
elif command -v shasum >/dev/null 2>&1; then
	sum() { shasum -a 256 "$1" | awk '{print $1}'; }
else
	die "need sha256sum or shasum"
fi

# Detect OS and architecture, mapped to cress's release matrix.
os=$(uname -s)
case "$os" in
Linux) os=linux ;;
Darwin) os=darwin ;;
*) die "unsupported OS: $os (cress ships linux and darwin)" ;;
esac
arch=$(uname -m)
case "$arch" in
x86_64 | amd64) arch=amd64 ;;
aarch64 | arm64) arch=arm64 ;;
*) die "unsupported arch: $arch (cress ships amd64 and arm64)" ;;
esac

# Resolve the latest tag if none was pinned. per_page=1 returns the newest
# release including prereleases, so this works before a stable tag exists.
#
# The API pretty-prints, so the separator is '": "' rather than '":"'. Matching
# with -n and p is what makes a parse failure safe: a pattern that does not
# match prints nothing, leaving TAG empty for the check below, where a
# substitution that does not match would pass the whole line through and build
# a URL out of it.
if [ -z "$TAG" ]; then
	log "resolving latest release..."
	TAG=$(dls "${API}/releases?per_page=1" \
		| tr ',' '\n' \
		| sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
		| head -1)
	[ -n "$TAG" ] || die "could not resolve the latest release; pass -t TAG"
fi

# GoReleaser strips the leading v from {{ .Version }} in the asset name.
ver=${TAG#v}
asset="${REPO}_${ver}_${os}_${arch}.tar.gz"
url="${BASE}/releases/download/${TAG}/${asset}"
sumurl="${BASE}/releases/download/${TAG}/checksums.txt"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM

log "downloading ${asset} (${TAG})..."
dlo "$url" "${tmp}/${asset}" || die "download failed: $url"

log "verifying checksum..."
dlo "$sumurl" "${tmp}/checksums.txt" || die "checksums download failed: $sumurl"
want=$(awk -v f="$asset" '$2 == f {print $1}' "${tmp}/checksums.txt")
[ -n "$want" ] || die "no checksum for ${asset} in checksums.txt"
got=$(sum "${tmp}/${asset}")
[ "$want" = "$got" ] || die "checksum mismatch: want ${want}, got ${got}"

# The checksum proves the download arrived intact. It cannot prove who published
# it, because it travels with the file it attests, so cosign verifies the
# signature over checksums.txt when it is available.
#
# Available, not required: this script's job is bootstrapping a machine that has
# nothing installed yet, and demanding cosign to install a static site generator
# would trade a real barrier for a threat most users are not facing. A mismatch
# is fatal; a missing cosign is a note. Pass CRESS_REQUIRE_SIGNATURE=1 to make
# the absence fatal too.
if command -v cosign >/dev/null 2>&1; then
	log "verifying signature..."
	if dlo "${sumurl}.sig" "${tmp}/checksums.txt.sig" &&
		dlo "${sumurl}.pem" "${tmp}/checksums.txt.pem"; then
		cosign verify-blob "${tmp}/checksums.txt" \
			--signature "${tmp}/checksums.txt.sig" \
			--certificate "${tmp}/checksums.txt.pem" \
			--certificate-identity-regexp "^https://github.com/${OWNER}/${REPO}/\.github/workflows/release\.yml@refs/tags/" \
			--certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
			>/dev/null 2>&1 || die "signature verification failed for checksums.txt"
	else
		log "note: this release is not signed; checksum verified only"
	fi
elif [ -n "${CRESS_REQUIRE_SIGNATURE:-}" ]; then
	die "CRESS_REQUIRE_SIGNATURE is set but cosign is not installed"
else
	log "note: cosign not found; checksum verified, signature not checked"
fi

log "extracting..."
tar -xzf "${tmp}/${asset}" -C "$tmp" "$REPO" || die "extract failed"
[ -f "${tmp}/${REPO}" ] || die "binary '${REPO}' not found in archive"

mkdir -p "$BINDIR" 2>/dev/null || true
[ -w "$BINDIR" ] || die "cannot write to ${BINDIR}; re-run with sudo, or pass -b DIR (e.g. -b \"\$HOME/.local/bin\")"
install -m 0755 "${tmp}/${REPO}" "${BINDIR}/${REPO}" 2>/dev/null \
	|| { cp "${tmp}/${REPO}" "${BINDIR}/${REPO}" && chmod 0755 "${BINDIR}/${REPO}"; } \
	|| die "install to ${BINDIR} failed"

log "installed ${REPO} ${TAG} -> ${BINDIR}/${REPO}"
case ":${PATH}:" in
*":${BINDIR}:"*) ;;
*) log "note: ${BINDIR} is not on your PATH" ;;
esac

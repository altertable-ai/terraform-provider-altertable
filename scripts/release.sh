#!/usr/bin/env bash
# Usage: ALTERTABLE_STAGING_TEST_API_KEY=<staging key> make release
set -euo pipefail

# Non-production endpoint the acceptance tests run against; they refuse the prod default.
readonly STAGING_API_URL="https://app.altertable.dev/rest/v1"
readonly SEMVER_RE='^v[0-9]+\.[0-9]+\.[0-9]+$'

die() { echo "error: $*" >&2; exit 1; }

is_version_greater() {
	[[ "$1" != "$2" ]] && [[ "$(printf '%s\n%s\n' "$1" "$2" | sort -V | tail -1)" == "$1" ]]
}

check_requirements() {
	: "${ALTERTABLE_STAGING_TEST_API_KEY:?set it to the staging API key so acceptance tests can run}"
	command -v gh >/dev/null || die "the GitHub CLI (gh) is required to trigger the release workflow"
	gh auth status >/dev/null 2>&1 || die "run 'gh auth login' first — the release workflow is triggered via gh"
	[[ -z "$(git status --porcelain)" ]] || die "working tree is dirty; commit or stash before releasing"
}

latest_release() {
	git fetch --tags --quiet
	git tag -l 'v*' --sort=-v:refname | head -1
}

prompt_version() {
	local latest="$1" version
	if [[ -n "$latest" ]]; then
		read -rp "Version to release (vX.Y.Z), latest is $latest: " version
	else
		echo "No previous releases found — this is the first release." >&2
		read -rp "Version to release (vX.Y.Z), e.g. v0.1.0: " version
	fi
	echo "$version"
}

validate_version() {
	local version="$1" latest="$2"
	[[ "$version" =~ $SEMVER_RE ]] || die "version must be vMAJOR.MINOR.PATCH (got '$version')"
	git rev-parse -q --verify "refs/tags/$version" >/dev/null && die "tag $version already exists"
	if [[ -n "$latest" ]]; then
		is_version_greater "$version" "$latest" || die "$version must be greater than the latest release $latest"
	fi
}

run_tests() {
	echo "==> Building"
	go build ./...
	echo "==> Unit tests"
	make test
	echo "==> Acceptance tests (against staging)"
	ALTERTABLE_API_KEY="$ALTERTABLE_STAGING_TEST_API_KEY" \
		ALTERTABLE_API_URL="$STAGING_API_URL" \
		make testacc
}

release() {
	local version="$1" confirm
	read -rp "All tests passed. Tag and release $version? [y/N] " confirm
	[[ "$confirm" == "y" || "$confirm" == "Y" ]] || die "aborted"
	git tag -a "$version" -m "Release $version"
	git push origin "$version"
	gh workflow run release.yml --ref "$version"
	echo "Pushed $version and triggered the release workflow. Watch it: gh run watch"
}

main() {
	check_requirements
	local latest version
	latest="$(latest_release)"
	version="$(prompt_version "$latest")"
	validate_version "$version" "$latest"
	run_tests
	release "$version"
}

main "$@"

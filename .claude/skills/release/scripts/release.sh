#!/usr/bin/env bash
# Publish a release: verify the repo is ready, run the quality checks, build
# the binaries, tag, and create the GitHub release with the CHANGELOG section
# of the version as its notes.
#
# Everything before the "Tag and publish" step is read-only, so a failure
# there leaves nothing to undo. If it fails after the tag was pushed, fix the
# cause and rerun only the failed command (e.g. `gh release create ...`);
# do not delete and re-push the tag.
set -euo pipefail

version=${1:?usage: release.sh vX.Y.Z}
[[ $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || {
  echo "error: version must look like vX.Y.Z, got '$version'" >&2
  exit 1
}

cd "$(git rev-parse --show-toplevel)"

# --- Preflight (read-only) --------------------------------------------------
fail() { echo "error: $1" >&2; exit 1; }

[[ -z $(git status --porcelain) ]] || fail "working tree is not clean; commit or stash first"
branch=$(git rev-parse --abbrev-ref HEAD)
[[ $branch == master ]] || fail "release from master, not '$branch'"
git fetch --quiet origin master
[[ $(git rev-parse HEAD) == $(git rev-parse origin/master) ]] \
  || fail "HEAD is not in sync with origin/master; push or pull first"
! git rev-parse --verify --quiet "refs/tags/$version" >/dev/null \
  || fail "tag $version already exists"
grep -q "^## $version\$" CHANGELOG.md \
  || fail "CHANGELOG.md has no '## $version' section; write it first"
grep -q "download/$version" README.md \
  || fail "README.md pinned-version example still points at an old version"
gh auth status >/dev/null 2>&1 || fail "gh is not authenticated; run 'gh auth login'"

make ci

# --- Release notes: the CHANGELOG section of this version, sans heading -----
notes=$(mktemp)
trap 'rm -f "$notes"' EXIT
awk -v v="$version" '$0 == "## " v {on=1; next} on && /^## / {exit} on' CHANGELOG.md > "$notes"
[[ -s $notes ]] || fail "the '## $version' CHANGELOG section is empty"

"$(dirname "$0")/build-dist.sh"

# --- Tag and publish --------------------------------------------------------
git tag -a "$version" -m "$version"
git push origin "$version"
gh release create "$version" dist/* --title "$version" --notes-file "$notes"

echo "released: https://github.com/benelog/uploader/releases/tag/$version"

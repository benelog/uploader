---
name: release
description: Release a new version of the uploader — write the CHANGELOG section, bump the README pinned-version example, then run the release script that verifies the repo, builds the per-platform binaries, tags, and publishes the GitHub release. Use this whenever the user asks to release, publish, tag, or ship a version (릴리스, 릴리즈, 배포), even if they only say "v2.2.0 내줘" or "release the current commits".
---

# Releasing the uploader

The mechanical part of a release is scripted; the judgment part is not.
Your job is the judgment part: pick the version, write the notes, then hand off to the script.

## 1. Pick the version

Semver, tagged as `vX.Y.Z`.
If the user named a version, use it.
Otherwise look at `git log <last-tag>..HEAD` and suggest: bump minor for new behavior or flags, patch for fixes only, major for breaking changes — and confirm with the user before publishing, since a release is public and permanent.

## 2. Write the CHANGELOG section

Add a `## vX.Y.Z` section at the top of `CHANGELOG.md` (below the `# Changelog` heading), summarizing `git log <last-tag>..HEAD` for a user of the tool, not a reader of the diff.
Follow the shape of the existing sections: a one-line summary sentence, then `### Highlights` with bold lead-ins, plus `### Notes` / `### Fixes` / `### Breaking changes` when they apply.
This file uses one sentence per line — no hard wrapping.
The section becomes the GitHub release notes verbatim, so write it to stand alone.

## 3. Bump the README pinned-version example

In `README.md`, the download instructions point at `releases/latest` and need no change; only the pinned example (`download/vX.Y.Z`) must be updated to the new version.
The script refuses to release while it still names an old version.

## 4. Commit and push

Commit the CHANGELOG and README changes and push to `origin master`.
The script releases exactly `origin/master`, so it refuses a dirty tree or an unpushed HEAD.

## 5. Run the script

    .claude/skills/release/scripts/release.sh vX.Y.Z

It re-verifies everything (clean tree, master in sync, tag free, CHANGELOG section present, README bumped, `gh` authenticated), runs `make ci`, extracts the CHANGELOG section as the release notes, cross-compiles the five binaries via `scripts/build-dist.sh`, then tags, pushes the tag, and creates the GitHub release with the binaries attached.

Everything before its "Tag and publish" step is read-only, so an early failure needs no cleanup — fix the reported problem and rerun.
If it fails *after* the tag was pushed, do not delete or re-push the tag: fix the cause and rerun just the failed command (usually `gh release create vX.Y.Z dist/* --title vX.Y.Z --notes-file <notes>`).

## 6. Report

Give the user the release URL and confirm the five assets are attached (`gh release view vX.Y.Z --json assets`).

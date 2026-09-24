---
name: release
description: Cut a steampipe CLI patch release from a v{maj}.{min}.x release branch - release turbot/steampipe-postgres-fdw first when it changed (version and changelog, tag, draft release, image publish) and bump FdwVersion, then steampipe's own changelog and release issue, the release workflow, verification and release notes, the merge-back, main and steampipe.io PRs, and the hand-over to the Pipes release.
---

# Release steampipe

Release branches are `v{maj}.{min}.x` (e.g. `v2.4.x`). A release is a tag cut from the branch head by the
`01 - Steampipe: Release` workflow. Never tag through the GitHub Releases UI: it creates the tag but
does not build binaries, publish, or update Homebrew.

Placeholders: `{x.y.z}` is the new version (e.g. `2.4.8`), `{prev}` the previous tag (e.g. `v2.4.7`,
from `gh release list --repo turbot/steampipe --limit 1 --exclude-pre-releases`), `{x}-{y}-{z}` and `{xyz}` the same version
dash-separated (`2-4-8`) and with no separators (`248`). `{f.x.y.z}` is the FDW version being released
(e.g. `2.2.8`) and `v{fmaj}.{fmin}.x` its release branch (e.g. `v2.2.x`).

Every PR you open below is opened under your own `gh` auth and needs a teammate's approval before it
merges. Exceptions: the `turbot/homebrew-tap` PR, which the workflow opens and merges itself, and PRs
into steampipe.io `main` or the FDW release branch, which require no review.

## 1. FDW first, only if the FDW changed

Skip to section 2 if `turbot/steampipe-postgres-fdw` hasn't changed since the `FdwVersion` in
`pkg/constants/db.go`. No workflow tags an FDW release; the tag is pushed by hand.

1. PR into the FDW release branch `v{fmaj}.{fmin}.x`, committed as `v{f.x.y.z}`: set
   `fdwVersion = "{f.x.y.z}"` in `version/version.go` and add a `## v{f.x.y.z} [YYYY-MM-DD]` entry to
   `CHANGELOG.md`. Both, before tagging: the binary reports `version.go`'s value.
2. Tag the branch head as it is on GitHub, without touching your checkout (`<fdw>` = your FDW clone):
   ```bash
   git -C <fdw> fetch origin && git -C <fdw> tag -a v{f.x.y.z} -m v{f.x.y.z} origin/v{fmaj}.{fmin}.x      && git -C <fdw> push origin v{f.x.y.z}
   ```
   `Build Draft Release` (`buildimage.yml`, on `v*` tags matching `vN.N.N[-suffix]`) builds four
   platform binaries into a **draft** release `v{f.x.y.z}`.
3. Publish the draft: `gh release edit v{f.x.y.z} --repo turbot/steampipe-postgres-fdw --draft=false`.
4. Push the image, which steampipe downloads (`ghcr.io/turbot/steampipe/fdw:{f.x.y.z}`, plus `:latest`
   for a version with no suffix):
   ```bash
   gh workflow run publish.yml --repo turbot/steampipe-postgres-fdw --ref develop -f release=v{f.x.y.z}
   ```
5. Verify: `docker manifest inspect ghcr.io/turbot/steampipe/fdw:{f.x.y.z}` returns a manifest.
6. FDW PRs: `v{fmaj}.{fmin}.x` into `develop`, titled `Merge branch 'v{fmaj}.{fmin}.x' into develop`, and
   into `main`, titled `Release steampipe-postgres-fdw v{f.x.y.z}` (the FDW repo asks for both each release).
7. On a branch off `v{maj}.{min}.x`, set `FdwVersion = "{f.x.y.z}"` in `pkg/constants/db.go` and open a PR
   into `v{maj}.{min}.x`. It can be the same PR as the section 2 changelog entry; it must merge before
   section 3, and only once step 5 passes.

## 2. Confirm what is being released

```bash
gh api 'repos/turbot/steampipe/compare/{prev}...v{maj}.{min}.x' \
  -q '.commits[] | "\(.sha[0:8]) \(.commit.message | split("\n")[0])"'
gh api 'repos/turbot/steampipe/contents/pkg/constants/db.go?ref=v{maj}.{min}.x' -q .content | base64 -d | grep FdwVersion
```

- Every intended fix is in the list. For a security fix, credit the commit that actually changed the
  dependency, not an adjacent PR.
- `CHANGELOG.md` on `v{maj}.{min}.x` has an entry for `v{x.y.z}` in the style of earlier entries, dated
  today, committed with the message `v{x.y.z}` (via a PR into the release branch).
- Open the release issue from `.github/ISSUE_TEMPLATE/release_issue.md`: title `Steampipe v{x.y.z}`,
  label `release`. Its checklist names older workflows; follow this skill's order and tick the items
  that still apply.

## 3. Dispatch the release workflow

```bash
gh workflow run 01-steampipe-release.yaml --repo turbot/steampipe --ref v{maj}.{min}.x \
  -f environment='Final (RC and final release)' -f version={x.y.z} -f confirmDevelop=false
sleep 15; gh run list --repo turbot/steampipe --workflow 01-steampipe-release.yaml --limit 3 --json databaseId,createdAt,headBranch
gh run watch --repo turbot/steampipe <run-id>
```

Take the run whose `createdAt` is after your dispatch; if none is, list again. `version` has no `v` prefix;
the workflow adds it. The `--ref` is the branch that gets tagged.
`Development (alpha)` / `Development (beta)` are for pre-release test builds only.

The `Release CLI` job creates the tag and GitHub release; later jobs open and merge the homebrew-tap PR
and dispatch `12 - Test: Linux Distros (Post-release)`.

## 4. Verify and publish the release notes

```bash
gh release view v{x.y.z} --repo turbot/steampipe
gh pr list --repo turbot/homebrew-tap --state merged --limit 5
gh run list --repo turbot/steampipe --workflow 12-test-post-release-linux-distros.yaml --limit 1
```

The release has its binaries, the homebrew-tap PR for `{x.y.z}` is merged, and the post-release test run
passed. The release is published with an empty body (goreleaser's changelog is disabled): edit it and
paste in the `CHANGELOG.md` entry for `v{x.y.z}`.

## 5. PRs

1. `v{maj}.{min}.x` into `develop`, titled `Merge branch 'v{maj}.{min}.x' into develop`. If the branches
   conflict (usually `go.mod`/`go.sum`), open it from a branch cut off `develop` that merges
   `origin/v{maj}.{min}.x` with the conflicts resolved (keep the higher version of each dependency), then `go mod tidy`.
2. `v{maj}.{min}.x` into `main`, titled `Release Steampipe v{x.y.z}`, label `release`. Never merge `main` into
   the release branch: if they conflict, open it from a branch cut off `v{maj}.{min}.x` that merges
   `origin/main`, keeping `main`'s workflow action pins. Body:
   ```
   ## Release Issue
   [Steampipe v{x.y.z}](<release issue URL>)

   ## Checklist
   - [ ] Confirmed that version has been correctly upgraded.
   ```
3. `turbot/steampipe.io`: branch `sp-{xyz}` off `main`, add
   `content/changelog/<YYYY>/<YYYYMMDD>-steampipe-cli-v{x}-{y}-{z}.md`:
   ```
   ---
   title: Steampipe CLI v{x.y.z} - <short summary>
   publishedAt: "<YYYY-MM-DD>T10:00:00"
   permalink: steampipe-cli-v{x}-{y}-{z}
   tags: cli
   ---
   ```
   Body matches the `CHANGELOG.md` entry. PR title `Steampipe CLI v{x.y.z}`, base `main`.
   Once merged, run the `Deploy steampipe.io` workflow from `main` and check the page loads.

## 6. Hand over

Give the new version (`steampipeCliVersion`) to whoever runs the Turbot Pipes release, tick the release
issue's checklist, and close it.

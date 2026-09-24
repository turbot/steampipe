---
name: release
description: Cut a steampipe CLI patch release from a v{maj}.{min}.x release branch - release turbot/steampipe-postgres-fdw first when it changed (changelog, tag, draft release, image publish) and bump FdwVersion, then steampipe's own changelog and release issue, the release workflow, verification and release notes, the merge-back, main and steampipe.io PRs, and the hand-over to the Pipes release.
---

# Release steampipe

Release branches are `v{maj}.{min}.x` (e.g. `v2.4.x`). A release is a tag cut from the branch head by the
`01 - Steampipe: Release` workflow. Never tag through the GitHub Releases UI: it creates the tag but
does not build binaries, publish, or update Homebrew.

Placeholders: `{x.y.z}` is the new version (e.g. `2.4.8`), `{prev}` the previous tag (e.g. `v2.4.7`,
from `gh release list --repo turbot/steampipe --limit 1`), `{x}-{y}-{z}` and `{xyz}` the same version
dash-separated (`2-4-8`) and with no separators (`248`). `{f.x.y.z}` is the FDW version being released
(e.g. `2.2.7`), distinct from steampipe's own `{x.y.z}`.

Every PR you open below is opened under your own `gh` auth and needs a teammate's approval before it
merges. The one exception is the `turbot/homebrew-tap` PR, which the workflow opens and merges itself.

## 1. FDW first, only if the FDW changed

Skip to step 2 if `turbot/steampipe-postgres-fdw` hasn't changed since the FDW version currently in
`pkg/constants/db.go`. No workflow tags an FDW release - every step below is manual.

1. On the FDW's own release branch (`v{fmaj}.{fmin}.x`, e.g. `v2.2.x`, reused across patches the same way
   steampipe's own release branch is), confirm `CHANGELOG.md` has an entry for `{f.x.y.z}` in the existing
   style (`## v{f.x.y.z} [YYYY-MM-DD]`), added via a PR into that branch.
2. Tag the branch head and push the tag - this is what triggers the build, nothing else does:
   ```bash
   cd ../steampipe-postgres-fdw && git checkout v{fmaj}.{fmin}.x && git pull
   git tag v{f.x.y.z} && git push origin v{f.x.y.z}
   ```
   `Build Draft Release` (`buildimage.yml`, triggered on any `v*` tag push) builds all four platform binaries
   and opens a **draft** GitHub release named `v{f.x.y.z}`.
3. Publish the draft: `gh release edit v{f.x.y.z} --repo turbot/steampipe-postgres-fdw --draft=false`.
4. Dispatch `Publish FDW Image` with that tag:
   ```bash
   gh workflow run publish.yml --repo turbot/steampipe-postgres-fdw -f release=v{f.x.y.z}
   ```
   It downloads the release's assets and pushes `ghcr.io/turbot/steampipe/fdw:{f.x.y.z}` (and `:latest`,
   unless `{f.x.y.z}` is an rc). Steampipe pulls the FDW from this image, not the GitHub release, so it must
   exist before step 2 below.
5. Verify the image landed: `docker manifest inspect ghcr.io/turbot/steampipe/fdw:{f.x.y.z}` (needs `docker`
   locally), or `gh api /orgs/turbot/packages/container/steampipe%2Ffdw/versions` if your token has
   `read:packages`.
6. Open the merge-back PR, `v{fmaj}.{fmin}.x` into `develop`, titled
   `Merge branch 'v{fmaj}.{fmin}.x' into develop`. A PR into `main` (`v{fmaj}.{fmin}.x` → `main`, titled
   `Release steampipe-postgres-fdw v{f.x.y.z}`) catches `main` up when it's fallen behind - it does not
   follow every patch.
7. On a branch off `v{maj}.{min}.x`, set `FdwVersion = "{f.x.y.z}"` in `pkg/constants/db.go` and open a
   PR into `v{maj}.{min}.x`. It can be the same PR as the step 2 changelog entry below; it must merge
   before step 3.

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
sleep 15; gh run list --repo turbot/steampipe --workflow 01-steampipe-release.yaml --limit 1 --json databaseId,createdAt,headBranch
gh run watch --repo turbot/steampipe <run-id>
```

Check `createdAt` and `headBranch` are your dispatch, not the previous release. `version` has no `v` prefix; the workflow adds it. The `--ref` is the branch that gets tagged.
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
2. `v{maj}.{min}.x` into `main`, titled `Release Steampipe v{x.y.z}`, label `release`, body:
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

---
name: release
description: Cut a steampipe CLI patch release from a v{maj}.{min}.x release branch - optional FDW bump, changelog and release issue, the release workflow, verification and release notes, the merge-back, main and steampipe.io PRs, and the hand-over to the Pipes release.
---

# Release steampipe

Release branches are `v{maj}.{min}.x` (e.g. `v2.4.x`). A release is a tag cut from the branch head by the
`01 - Steampipe: Release` workflow. Never tag through the GitHub Releases UI: it creates the tag but
does not build binaries, publish, or update Homebrew.

Placeholders: `{x.y.z}` is the new version (e.g. `2.4.8`), `{prev}` the previous tag (e.g. `v2.4.7`,
from `gh release list --repo turbot/steampipe --limit 1`), `{x}-{y}-{z}` and `{xyz}` the same version
dash-separated (`2-4-8`) and with no separators (`248`).

Every PR you open below is opened under your own `gh` auth and needs a teammate's approval before it
merges. The one exception is the `turbot/homebrew-tap` PR, which the workflow opens and merges itself.

## 1. FDW first, only if the FDW changed

1. Release `turbot/steampipe-postgres-fdw`, then run its `Publish FDW Image` workflow with that release
   tag. Steampipe downloads the FDW from `ghcr.io/turbot/steampipe/fdw:<version>`, not from the GitHub
   release, so the image must exist before the CLI ships.
2. On a branch off `v{maj}.{min}.x`, set `FdwVersion = "<fdw version>"` in `pkg/constants/db.go` and open a
   PR into `v{maj}.{min}.x`. It can be the same PR as the step 2 changelog entry; it must merge before step 3.

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
gh run list --repo turbot/steampipe --workflow 01-steampipe-release.yaml --limit 1 --json databaseId -q '.[0].databaseId'
gh run watch --repo turbot/steampipe <run-id>
```

`version` has no `v` prefix; the workflow adds it. The `--ref` is the branch that gets tagged.
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

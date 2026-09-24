---
name: release
description: Cut a steampipe CLI patch release from a v{maj}.{min}.x release branch - optional FDW bump, changelog, the release workflow, verification, and the merge-back, main and steampipe.io PRs.
---

# Release steampipe

Release branches are `v{maj}.{min}.x` (e.g. `v2.4.x`). A release is a tag cut from the branch head by the
`01 - Steampipe: Release` workflow. Never tag through the GitHub Releases UI: it creates the tag but
does not build binaries, publish, or update Homebrew.

Every PR below is opened under your own `gh` auth and needs a teammate's approval before it merges.

## 1. FDW first, only if the FDW changed

1. Release `turbot/steampipe-postgres-fdw` and wait for its GitHub release to publish.
2. On a branch off `v{maj}.{min}.x`, set `FdwVersion = "X.Y.Z"` in `pkg/constants/db.go` and open a PR
   into `v{maj}.{min}.x`. It must merge before step 3.

## 2. Confirm what is being released

```bash
gh api repos/turbot/steampipe/compare/v{prev}...v{maj}.{min}.x \
  -q '.commits[] | "\(.sha[0:8]) \(.commit.message | split("\n")[0])"'
gh api repos/turbot/steampipe/contents/pkg/constants/db.go?ref=v{maj}.{min}.x -q .content | base64 -d | grep FdwVersion
```

- Every intended fix is in the list. For a security fix, credit the commit that actually changed the
  dependency, not an adjacent PR.
- `CHANGELOG.md` on `v{maj}.{min}.x` has an entry for `v{x.y.z}` in the style of earlier entries, dated
  today, committed with the message `v{x.y.z}` (via a PR into the release branch).
- Open the release issue from `.github/ISSUE_TEMPLATE/release_issue.md`: title `Steampipe v{x.y.z}`,
  label `release`.

## 3. Dispatch the release workflow

```bash
gh workflow run 01-steampipe-release.yaml --repo turbot/steampipe --ref v{maj}.{min}.x \
  -f environment='Final (RC and final release)' -f version={x.y.z} -f confirmDevelop=false
```

`version` has no `v` prefix; the workflow adds it. The `--ref` is the branch that gets tagged.
`Development (alpha)` / `Development (beta)` are for pre-release test builds only.

Watch it with `gh run watch --repo turbot/steampipe <run-id>`. `build_and_release_cli` creates the tag
and GitHub release; later jobs open and merge the `turbot/homebrew-tap` PR and start smoke tests.

## 4. Verify

```bash
gh release view v{x.y.z} --repo turbot/steampipe
gh pr list --repo turbot/homebrew-tap --state merged --limit 5
```

The release has its binaries and the homebrew-tap PR for `{x.y.z}` merged. If the generated release notes
are wrong, edit the release to match `CHANGELOG.md`.

## 5. PRs

1. `v{maj}.{min}.x` into `develop`, titled `Merge branch 'v{maj}.{min}.x' into develop`.
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
   Body is the `CHANGELOG.md` entry minus CI-only items. PR title `Steampipe CLI v{x.y.z}`, base `main`.
   Once merged, run the `Deploy steampipe.io` workflow from `main` and check the page loads.

## 6. Hand over

Give the new version (`steampipeCliVersion`) to whoever runs the Turbot Pipes release, tick the release
issue's checklist, and close it.

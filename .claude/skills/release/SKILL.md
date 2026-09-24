---
name: release
description: Cut a steampipe CLI release — FDW version bump, release-branch verification, the release workflow dispatch, and the follow-up PRs (merge-back, main, steampipe.io changelog). Use whenever the user asks to cut, tag, or ship a steampipe release, or to bump the FDW version ahead of one. Every PR opens under the operator's own `gh` auth and needs a teammate's approval before merge — this skill never merges a PR itself.
---

# Steampipe release

Six steps, in order. Skip step 1 if the FDW isn't changing.

## 1. Bump the FDW version (only if FDW changed)

Tag the FDW release first (`turbot/steampipe-postgres-fdw`) and wait for the artifact to publish. Then bump the constant on the steampipe release branch (`v{maj}.{min}.x`):

- Edit `pkg/constants/db.go` — `FdwVersion = "X.Y.Z"`.
- Land it via a PR into `v{maj}.{min}.x`. Needs a teammate's approval before merge.

## 2. Confirm the release branch contents

Before dispatching anything, confirm what's actually on the branch:

```
gh api repos/turbot/steampipe/compare/v{prev}...v{maj}.{min}.x
```

Read the commit list — don't assume. Also confirm the `CHANGELOG.md` entry for this version is present and matches the commits.

## 3. Dispatch the release workflow

Workflow: `01 - Steampipe: Release` (`.github/workflows/01-steampipe-release.yaml`).

```
gh workflow run 01-steampipe-release.yaml \
  --ref v{maj}.{min}.x \
  -f environment='Final (RC and final release)' \
  -f version=<x.y.z> \
  -f confirmDevelop=false
```

- `environment`: `Final (RC and final release)` for a real release (the `Development (alpha)` / `Development (beta)` options are for pre-release testing).
- `version`: patch version **without** the `v` prefix, e.g. `2.4.1`. The workflow prepends `v` when tagging.
- `confirmDevelop`: `false` — only `true` when the branch selector is `develop`, which is not the normal release path.

Run this under the operator's own `gh` auth. The workflow tags, builds and publishes the release, and opens the homebrew-tap PR.

## 4. Verify the release published

```
gh release view v{x.y.z}
```

Also confirm the homebrew-tap PR merged.

## 5. Open the follow-up PRs

All three are opened under the operator's own `gh` auth and each needs a teammate's approval before merge:

1. **Merge-back into `develop`** — title `Merge branch '<branchname>' into develop`.
2. **Release into `main`** — title `Release Steampipe v<version>`, body:
   ```
   ## Release Issue
   [Steampipe v<version>](link-to-release-issue)

   ## Checklist
   - [ ] Confirmed that version has been correctly upgraded.
   ```
3. **steampipe.io changelog** — new file at `content/changelog/<YYYY>/<YYYYMMDD>-steampipe-cli-v<X>-<Y>-<Z>.md`, frontmatter:
   ```yaml
   ---
   title: Steampipe CLI v<version> - <short summary>
   publishedAt: "<YYYY-MM-DD>T10:00:00"
   permalink: steampipe-cli-v<version-with-dashes>
   tags: cli
   ---
   ```
   PR title: `Steampipe CLI v<version>`.

## 6. Hand off

Give `steampipeCliVersion` to whoever runs the next Pipes workspace release — the workspace image bundles steampipe.

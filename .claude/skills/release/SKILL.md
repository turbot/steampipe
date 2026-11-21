---
name: release
description: Release steampipe through the shared release workflows - the FDW first when it changed, the prep file, the cut and the release PR, the approval, watching the ship (tag, Homebrew, steampipe.io page, distro test, merge-back), a re-ship, and the hand-over to the Pipes release.
---

# Release steampipe

steampipe releases through Turbot's shared release workflows (turbot/shipyard-workflows, called at `v0`). Nobody types a
version, pushes a tag, edits `CHANGELOG.md` or opens a Homebrew or steampipe.io PR by hand. Every PR into `develop` that
ships carries a `Changelog:` line, and those lines are the release's changelog.

The Release carries the darwin `.zip` and linux `.tar.gz` archives, the linux `.deb` and `.rpm` packages and
`checksums.txt`, with the same names as every release before (`steampipe_<os>_<arch>.<ext>`): the Homebrew tap,
`scripts/install.sh` and the setup action fetch them by name.

## 1. FDW first, only if the FDW changed

Skip to section 2 if `turbot/steampipe-postgres-fdw` has nothing new since the `FdwVersion` in
`pkg/constants/db.go`: `gh api repos/turbot/steampipe-postgres-fdw/compare/v<FdwVersion>...v{fmaj}.{fmin}.x -q .ahead_by`.
`{f.x.y.z}` is the FDW version being released (e.g. `2.2.8`) and `v{fmaj}.{fmin}.x` its release branch (e.g. `v2.2.x`).
The FDW repo is not on the shared release workflows yet: its tag is still pushed by hand, there.

1. PR into the FDW release branch `v{fmaj}.{fmin}.x`, committed as `v{f.x.y.z}`: set
   `fdwVersion = "{f.x.y.z}"` in `version/version.go` and add a `## v{f.x.y.z} [YYYY-MM-DD]` entry to
   `CHANGELOG.md`. Both, before tagging: the binary reports `version.go`'s value.
2. Tag the branch head as it is on GitHub, without touching your checkout (`<fdw>` = your FDW clone):
   ```bash
   git -C <fdw> fetch origin && git -C <fdw> tag -a v{f.x.y.z} -m v{f.x.y.z} origin/v{fmaj}.{fmin}.x && git -C <fdw> push origin v{f.x.y.z}
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
7. Set `FdwVersion = "{f.x.y.z}"` in `pkg/constants/db.go` by a PR into steampipe's `develop`, with a `Changelog:`
   line, only once step 5 passes. It must merge before the prep in section 2 reads `develop`.

## 2. Start and prep

`release-start.yml` files a prep job (an issue labelled `agent:release`) on Tuesdays at 09:00 UTC when a shipped change
waits, or when run by hand on `develop`. Dependency bumps alone start nothing, unless one fixes an advisory.

Write the prep file with the skill the prep job names, for the commit it names. Then a person runs
`prep.yml` with the file's content, and `release.yml` on `develop` with `commit` (the full id) and `prep-run` (the prep run's
id). The cut makes `release/<version>` with the CHANGELOG section and opens the release PR into `main`.

## 3. The release PR

`release-pr.yml` tests the candidate (`go test ./...`), builds a snapshot, installs it and writes the record. It also runs
`10 - Test: Linting` and `11 - Test: Acceptance` on the PR's commit, the checks `develop` requires. With a SHIP record, the
record marks the PR ready for review and starts `ask.yml`, which asks in #steampipe within a minute. `main` accepts the
merge only when the record's `shipyard/release` status passes. One approver of `turbot/pipelings-release-approvers` who
drove none of its changes approves it. Merge with a merge commit.

## 4. Watch the ship

`ship.yml` runs on the merge:

```
ship (build at the merge) ──► tag (tag, GitHub Release with the archives and packages, merge-back PR into develop)
                                ├──► homebrew (turbot/homebrew-tap: steampipe.rb and steampipe@<version>.rb)
                                ├──► page (steampipe.io: content/changelog/steampipe-cli-v<x-y-z>.md)
                                └──► distros (12 - Test: Linux Distros (Post-release))
```

```bash
gh run list --repo turbot/steampipe --workflow ship.yml --limit 1
gh release view v{x.y.z} --repo turbot/steampipe
brew update && brew install turbot/tap/steampipe && steampipe --version
```

It is done when:

- Homebrew installs the new version and `steampipe --version` says it;
- `scripts/install.sh` installs it in a clean Linux container;
- the `.deb` and `.rpm` files are on the Release;
- the page is live at `https://steampipe.io/changelog/steampipe-cli-v{x}-{y}-{z}`.

A failed `homebrew` or `page` job is finished by running it again; it writes nothing twice. The site deploys on each push
to its `main`; if the page job passed and the page is still not live, run the site's `Deploy steampipe.io` workflow from
`main`.

**A release merged but not shipped** (the ship job itself is broken): `main` takes only release PRs, so run `release.yml`
on `develop` with `reship` (the version), `commit` (the tip of `main`) and `fixes` (the merge commits of the fix on
`develop`). A fix that changes `.github/workflows/` cannot go in `fixes`: the cut says how to pick it by hand.

## 5. Hand over

Tell whoever runs the Turbot Pipes release that the new version is live, with its tag (`v{x.y.z}`). Pipes pins the CLI by
that tag (`steampipeCliVersion`), and the new version ships in the next Pipes release.

---
name: Steampipe Release
about: Steampipe Release
title: "Steampipe v<INSERT_VERSION_HERE>"
labels: release
---

#### Changelog

[Steampipe v<INSERT_VERSION_HERE> Changelog](https://github.com/turbot/steampipe/blob/v<INSERT_VERSION_HERE>/CHANGELOG.md)

## Checklist

### Pre-release checks

- [ ] All acceptance tests pass in `steampipe` release PR
- [ ] Update check is working
- [ ] Steampipe version is correct
- [ ] Steampipe Changelog updated and reviewed

### Release Steampipe

- [ ] Merge the release PR
- [ ] Trigger the `Steampipe CLI Release` workflow. This will create the release build.
- [ ] Trigger the `Publish and Update Brew` workflow. This will update the brew formula.

### Post-release checks

- [ ] Update Changelog in the Release page (copy and paste from CHANGELOG.md)
- [ ] Test Linux install script
- [ ] Test Homebrew install
- [ ] Release branch merged to `develop`
- [ ] Raise Changelog update to `steampipe.io`, get it reviewed.
- [ ] Merge Changelog update to `steampipe.io`.
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

- [ ] The release PR's page shows its checks passed and what changes
- [ ] Update check is working

### Release Steampipe

- [ ] A pipelings release approver approves the release PR and merges it. The merge builds, tags, publishes the GitHub Release with its notes, archives and Linux packages, updates Homebrew, writes the changelog page on steampipe.io, starts the Linux distro test and merges `main` back into `develop`.

### Post-release checks

- [ ] Test Linux install script
- [ ] Test Homebrew install
- [ ] Check the `.deb` and `.rpm` files are on the Release
- [ ] Check the changelog page is live on steampipe.io
- [ ] Tell the Turbot Pipes release the new tag: Pipes pins the CLI by it

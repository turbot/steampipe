---
description: Check and fix Dependabot security vulnerabilities
allowed-tools: Bash(gh api:*), Bash(gh release:*), Bash(yarn:*), Bash(go:*), Bash(make:*), Bash(git branch:*), Bash(git checkout:*), Bash(git log:*), Bash(git add:*), Bash(gh pr create:*), Skill(commit), Skill(push)
---

Remediate security vulnerabilities reported by Dependabot. Follow these steps:

## Step 1: Determine the base branch

1. Get the repository owner/name from `gh repo view --json owner,name`
2. Get the latest release: `gh release list --limit 1`
3. Derive the release branch by replacing the patch version with `x` (e.g., `v1.4.2` → `v1.4.x`)
4. Verify the branch exists: `git branch -r | grep <branch>`

**Ask the user**: "The latest release is `{tag}` and the release branch is `{branch}`. Should I use this as the base branch, or use `develop` instead?"

## Step 2: Check for vulnerabilities

1. Run `gh api repos/{owner}/{repo}/dependabot/alerts --paginate` to list open alerts
2. Filter by state=open and sort by severity (critical/high first)
3. Present a summary table: Alert #, Package, Ecosystem, Severity, CVE, Fix Version

**Ask the user**: Which vulnerabilities to fix (all high, specific ones, all)?

## Step 3: Apply fixes

### For npm dependencies:
1. Check current version: `yarn why <package>`
2. Check existing patterns: `git log --oneline --grep="vulnerab"`
3. Direct deps → update version in `package.json`
4. Transitive deps → add to `resolutions` in `package.json`
5. Run `yarn install`
6. Verify: `yarn why <package>`

### For Go dependencies:
1. Run `go get <package>@<version>`
2. Run `go mod tidy`

**Important**: For major version changes, ask user confirmation first.

## Step 4: Build and test

1. Go: Run `make` and `go test ./...`
2. npm: Run `yarn build` in the UI directory
3. Report failures before proceeding

## Step 5: Commit, push, and create PR

1. Checkout base branch and create: `fix/vulnerability-updates-{base-branch}`
2. Stage relevant files only (package.json, yarn.lock, go.mod, go.sum)
3. Use `/commit` with message listing packages, versions, and CVEs
4. Use `/push` to push the branch
5. Create PR: `gh pr create --base {base-branch}` with summary of fixes

Return the PR URL when done.

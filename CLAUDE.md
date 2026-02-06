# Release Process

Follow these steps in order to perform a release:

## 1. Changelog
- Draft a changelog entry in `CHANGELOG.md` matching the style of existing entries.
- Use today's date and the next patch version.

## 2. Commit
- Commit message for release changelog changes should be the version number, e.g. `v2.3.5`.

## 3. Release Issue
- Use the `.github/ISSUE_TEMPLATE/release_issue.md` template.
- Title: `Steampipe v<version>`, label: `release`.

## 4. PRs
1. **Against `develop`**: Title should be `Merge branch '<branchname>' into develop`.
2. **Against `main`**: Title should be `Release Steampipe v<version>`.
   - Body format:
     ```
     ## Release Issue
     [Steampipe v<version>](link-to-release-issue)

     ## Checklist
     - [ ] Confirmed that version has been correctly upgraded.
     ```
   - Tag the release issue to the PR (add `release` label).

## 5. steampipe.io Changelog
- Create a changelog PR in the `turbot/steampipe.io` repo.
- Branch off `main`, branch name: `sp-<version without dots>` (e.g. `sp-235`).
- Add a file at `content/changelog/<year>/<YYYYMMDD>-steampipe-cli-v<version-with-dashes>.md`.
- Frontmatter format:
  ```
  ---
  title: Steampipe CLI v<version> - <short summary>
  publishedAt: "<YYYY-MM-DD>T10:00:00"
  permalink: steampipe-cli-v<version-with-dashes>
  tags: cli
  ---
  ```
- Body should match the changelog content from `CHANGELOG.md`.
- PR title: `Steampipe CLI v<version>`, base: `main`.

## 6. Deploy steampipe.io
- After the steampipe.io changelog PR is merged, trigger the `Deploy steampipe.io` workflow in `turbot/steampipe.io` from `main`.

## 7. Close Release Issue
- Check off all items in the release issue checklist as steps are completed.
- Close the release issue once all steps are done.

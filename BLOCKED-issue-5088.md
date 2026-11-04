## Kind: design

# BLOCKED: #5088 add-to-project / add-to-project red on develop

## What I was doing
Looking for an in-repo change that makes `add-to-project / add-to-project` pass at develop's tip.

## What I found
- `.github/workflows/31-add-issues-to-pipeling-issue-tracker.yaml` only forwards `issue_number` and `repository` to
  `turbot/steampipe-workflows/.github/workflows/assign-issue-to-pipeling-issue-tracker.yml@main` with `secrets: inherit`.
  Nothing in this repo sets the project number, the action version or `fields`.
- That reusable workflow (read via the API at main) runs `actions/add-to-project@244f685b...` (v1.0.2) with
  `project-url: https://github.com/orgs/turbot/projects/31`, `github-token: secrets.GH_PROJECT_PAT`, `fields: '{"Status": "Todo"}'`.
- The failure is `Could not resolve to a ProjectV2 with the number 31.` The `fields` input is not valid for v1.0.2
  (valid: project-url, github-token, labeled, label-operator); that is only a warning, not the failure.
- The check has failed on every recent `issues` run (`gh run list`).
- From the agent account, `organization(login:"turbot"){projectV2(number:31)}` returns NOT_FOUND. That fits a deleted or renumbered project, or a
  `GH_PROJECT_PAT` that lost access, but the agent account cannot tell which.

## Ruled out
- A change to the caller file in this repo (it holds none of the failing inputs).
- Guessing a new project number.

## Next step (needs a person with org access)
1. Check that org project turbot/31 still exists (or was renumbered) and that the `GH_PROJECT_PAT` secret can read and write it; renew or re-scope the PAT if it expired.
2. If the project was renumbered, update `project-url` in turbot/steampipe-workflows; also drop `fields` or move to an add-to-project version that supports it.
3. Re-run the failed job; the check then passes at develop's tip. Consider tracking this in turbot/steampipe-workflows rather than here.

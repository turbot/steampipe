---
name: soc2-remediation
description: >-
  Weekly SOC2 vulnerability remediation for this repo. Pulls the Vanta findings
  snapshot from turbot/soc2, reconciles it with Dependabot, classifies each
  finding as mergeable / manual fix / no-fix, remediates what has a fix, and
  prepares snoozes for what does not. Use whenever the operator mentions Vanta,
  CVE remediation, Dependabot security alerts, vulnerability SLAs, or the
  weekly SOC2 run.
---

# SOC2 vulnerability remediation

Read-only toward Vanta — never writes to it and never merges anything without
the operator, who approves manual-bump PRs and runs the snooze batch.

## Snapshot

Pull the weekly findings snapshot (no manual paste needed):

```bash
gh api -H "Accept: application/vnd.github.raw" \
  repos/turbot/soc2/contents/monitoring/state/vulnerability-state.json > /tmp/vuln-state.json
```

Top-level `lastUpdated` + `vulnerabilities[]`, each with `externalId`, `severity`,
`status` (`OVERDUE` / `DUE_SOON` / `OK`), `remediateByDate`, `packageName` (with
the vulnerable range), `isFixable`, `repo`, `integration`, `id`. It's a weekly
snapshot and lags Dependabot — state the `lastUpdated` date.

Default scope is the current repo (`gh repo view --json name -q .name`); accept
a list of repos as an argument instead. Filter and sort by SLA:

```bash
jq -r --arg repo steampipe '
  .vulnerabilities | map(select(.repo == $repo))
  | sort_by(.remediateByDate) | .[]
  | "\(.status)\t\(.remediateByDate[:10])\t\(.severity)\t\(.repo)\t\(.externalId)\t\(.packageName)"
' /tmp/vuln-state.json
```

## Reconcile

Always union with the live Dependabot alerts — this is not optional:

```bash
gh api repos/turbot/<repo>/dependabot/alerts --paginate \
  --jq '.[]|select(.state=="open")|"\(.security_advisory.severity) \(.dependency.package.name) \(.security_vulnerability.vulnerable_version_range) fix=\(.security_vulnerability.first_patched_version.identifier // "none")"' | sort -u
```

Anything GitHub shows that the snapshot doesn't still needs fixing; it just
can't be snoozed until Vanta's next scan creates a row for it.

Map each finding to its Dependabot PR by package name + the fix version implied
by the vulnerable range:

```bash
gh pr list --repo turbot/<repo> --author app/dependabot --state open --limit 60 \
  --json number,title -q '.[] | "#\(.number)  \(.title)"'
```

## Classify

- **A — has a PR.** A Dependabot PR exists. The common case.
- **B — no PR, but a fix exists upstream.** Deeply nested transitive deps
  Dependabot can't raise a direct PR for. Not a snooze — a fix exists.
- **C — no fix.** No installable fixed version exists for the module in use.

Before calling C, check two traps: a **renamed module** (the fix may live
under a different import path the same project publishes), and a
**range with no lower bound** (e.g. `<= 5.0.7` includes every older major, so
a same-major pin can't clear it).

## Remediate

**Bucket A:** `gh pr comment <n> --repo turbot/<r> --body "@dependabot rebase"`,
confirm CI green (`gh pr checks <n>`), approve (`gh pr review <n> --approve`),
squash-merge (`gh pr merge <n> --squash --delete-branch`).

**Bucket B:** manual bump. First decide the base: `gh release list --limit 1`,
derive the release branch by replacing the patch version with `x`; ask whether
the fix targets that branch or `develop`. Go: `go get <module>@<version>` then
`go mod tidy`. npm: bump the parent, or pin via a `resolutions`/`overrides`
entry within the current major.

After any bump, build and test before opening a PR: `make` and `go test ./...`;
add `yarn build` in the UI directory only when the repo has one. Report
failures before proceeding. Bucket B PRs go under the operator's own `gh` auth
and need a teammate's approval — the author can't self-approve; bucket A PRs
the operator approves directly.

## Verify

A merged bump is not proof the finding cleared.

npm, only in a repo with a `yarn.lock` (check for `ui/` first — not every repo
has one): the lockfile must hold exactly one resolved version —
`grep -A1 '"<pkg>@npm' yarn.lock | grep version | sort -u`. Two lines means a
peer dependency or unused package still pins the old copy.

For every finding: re-run the Dependabot alert query and confirm the open
count dropped. Merging to `develop` clears the finding — Dependabot reads the
default branch, so no release or tag is needed.

## Snooze

For bucket C, write a short paste-ready justification: why no fix exists, and
why it's safe to wait (not reachable / EOL). For 3 or more items, also write a
batch file of `{id, reason, until, autoReopen}` for:

```bash
npm run auto-snooze -- --dry-run --batch <path>.json
```

Run from `turbot/soc2/monitoring`; needs `VANTA_CLIENT_ID` /
`VANTA_CLIENT_SECRET`, which only the operator holds. This skill prepares the
batch — it never runs the write itself. Recommend a reactivation date ~90 days
out with auto-reopen-on-fix enabled; not indefinite, not short.

## Output

Produce a plan sorted by SLA due date (nearest first):

1. **Merges** — per repo, PR number, what it fixes, who approves.
2. **Manual gaps** — package, repo, that a fix exists, the bump needed.
3. **Snoozes** — the finding, the justification, the reactivation date.

Lead with whatever is due within the week; note the rest as having runway.

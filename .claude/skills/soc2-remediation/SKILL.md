---
name: soc2-remediation
description: >-
  Weekly SOC2 vulnerability remediation for one or more pipeling repos (default:
  the current repo). Pulls the Vanta findings snapshot from turbot/soc2,
  reconciles it with Dependabot, classifies each finding as mergeable / manual
  fix / no-fix, remediates what has a fix, and prepares snoozes for what does
  not. Use whenever the operator mentions Vanta, CVE remediation, Dependabot
  security alerts, vulnerability SLAs, or the weekly SOC2 run.
---

# SOC2 vulnerability remediation

Read-only toward Vanta. Present the plan (see Output) and let the operator choose
what to fix before any merge or bump; confirm each merge. The operator runs the snooze batch.

## Snapshot

```bash
gh api -H "Accept: application/vnd.github.raw" \
  repos/turbot/soc2/contents/monitoring/state/vulnerability-state.json > /tmp/vuln-state.json
```

Top level: `lastUpdated`, `vulnerabilities[]` and `nowFixableSnoozedItems[]` (snoozed findings
that now have a fix — treat them as open). Each row has `id` (Vanta's id), `externalId` (CVE/GHSA),
`severity`, `status` (`OVERDUE` / `DUE_SOON` / `OK` / `NO_SLA`), `remediateByDate` (null when no
SLA), `packageName` (ecosystem prefix + vulnerable range, e.g. `go-go.opentelemetry.io/otel/sdk
>= 1.5.0, <= 1.44.0` — strip the ecosystem prefix to match Dependabot), `isFixable`, `repo`,
`integration`. One CVE can have several rows (one per affected package or path). The snapshot is
weekly and lags Dependabot — state its `lastUpdated` date.

Default scope is the current repo (`gh repo view --json name -q .name`); accept a list of repos as
an argument instead. Filter and sort by SLA:

```bash
jq -r --arg repo steampipe '
  (.vulnerabilities + .nowFixableSnoozedItems) | map(select(.repo == $repo))
  | sort_by(.remediateByDate // "9999") | .[]
  | "\(.status)\t\((.remediateByDate // "none")[:10])\t\(.severity)\t\(.externalId)\t\(.packageName)"
' /tmp/vuln-state.json
```

## Reconcile

Always union with the live Dependabot alerts:

```bash
gh api repos/turbot/<repo>/dependabot/alerts --paginate \
  --jq '.[]|select(.state=="open")|"\(.security_advisory.severity) \(.dependency.package.name) \(.security_vulnerability.vulnerable_version_range) fix=\(.security_vulnerability.first_patched_version.identifier // "none")"' | sort -u
gh pr list --repo turbot/<repo> --author app/dependabot --state open --limit 60 \
  --json number,title,baseRefName -q '.[] | "#\(.number) \(.baseRefName) \(.title)"'
```

Anything GitHub shows that the snapshot doesn't still needs fixing; it just can't be snoozed until
Vanta's next scan creates a row for it. Map each finding to its PR by package name and the fix
version implied by the range.

## Classify

- **A — has a PR.** An open Dependabot PR fixes it.
- **B — no PR, but a fix exists.** Typically a nested transitive dependency. Not a snooze.
- **C — no fix.** No installable fixed version exists for the module in use. A row with
  `isFixable: true`, or an alert with a `fix=` version, is never C.

Before calling C, check two traps: a **renamed module** (the fix may live under a different import
path the same project publishes), and a **range with no lower bound** (e.g. `<= 5.0.7` includes
every older major, so a same-major pin can't clear it).

## Remediate

Use `--repo turbot/<r>` on every `gh pr` command; the PR number alone resolves against the current
checkout.

**Bucket A:** `gh pr comment <n> --repo turbot/<r> --body "@dependabot rebase"`, then once CI is
green (`gh pr checks <n> --repo turbot/<r>`): `gh pr review <n> --repo turbot/<r> --approve` and
`gh pr merge <n> --repo turbot/<r> --squash --delete-branch`.

**Bucket B:** manual bump on a branch off `develop` — Dependabot and Vanta read the default branch,
so only a fix on `develop` clears the finding. Go: `go get <module>@<version>` then `go mod tidy`.
npm: bump the parent, or pin via `resolutions`/`overrides` within the current major, then
`yarn install` and `yarn why <pkg>`. Ask the operator before any major-version bump. PR title names
package, version and advisory, e.g. `Bump github.com/jackc/pgx/v5 to v5.9.2 (CVE-2026-41889)`.

If the fix must also ship in a release (`gh release list --repo turbot/<r> --limit 1`, release
branch = patch replaced by `x`), open a second PR into that branch titled `Backport: <title>`. Redo
the bump there with `go get` rather than cherry-picking; `go.mod` usually differs.

After any bump, build and test before opening the PR: `go build ./...` and `go test ./...` (plus
`yarn build` in the UI directory when the repo has one). Report failures before proceeding. Manual
PRs go under the operator's own `gh` auth and need a teammate's approval.

## Verify

A merged bump is not proof the finding cleared.

npm, only in a repo with a `yarn.lock` (not every repo has a `ui/`): the lockfile must hold exactly
one resolved version — `grep -A1 '"<pkg>@npm' yarn.lock | grep version | sort -u`. Two lines means
a peer dependency or unused package still pins the old copy.

For every finding: re-run the Dependabot alert query and confirm the open count dropped. For
`integration: github` rows, merging to `develop` clears the finding with no release. Findings
against a published image (e.g. the Pipes workspace image) clear only once a release carrying the
bump ships.

## Snooze

For bucket C, write a short paste-ready justification: why no fix exists, and why it's safe to wait
(not reachable / EOL). Snoozes are per snapshot row, so cover every `id` for the CVE. For 3 or more
rows, write a batch file of `{id, reason, until, autoReopen}` — `id` is Vanta's `id`, not the CVE;
`until` is ISO 8601. Dry-run it from `turbot/soc2/monitoring`:

```bash
npm run auto-snooze -- --dry-run --batch <path>.json
```

The dry run accepts unknown ids, so check each against the snapshot. The operator runs the real
write (`SNOOZE_INVOKED_BY=<their name> npm run auto-snooze -- --batch <path>.json`; single row:
`--id <id> --reason <text> --until <ISO>`) with `VANTA_CLIENT_ID` / `VANTA_CLIENT_SECRET`, which
only they hold; this skill never runs it. Recommend a reactivation date ~90 days out with
auto-reopen-on-fix enabled; not indefinite, not short.

## Output

A plan sorted by SLA due date (nearest first), for the operator to choose from:

1. **Merges** — per repo, PR number, what it fixes, who approves.
2. **Manual gaps** — package, repo, the fixed version, the bump needed, whether a backport is needed.
3. **Snoozes** — the finding, its row ids, the justification, the reactivation date.

Lead with whatever is due within the week; note the rest as having runway.

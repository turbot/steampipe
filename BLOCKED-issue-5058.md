# BLOCKED: issue #5058

## Kind: environment

The change is implemented, self-reviewed and pushed to `fix/issue-5058-connection-state-save`
(no PR yet). The one owed check left is `check:acceptance`, which is recorded `unmet: steampipe-ports`
on this worker: another run's `tests/acceptance/run-local.sh` (pid 3377 and its child, 12+ minutes
in when I stopped waiting) holds the machine's acceptance runner, so the `steampipe-ports` need
cannot be met. The manifest says this is unmet, not red, and a suite of that length does not fit
what is left of this run.

## Done and recorded at HEAD

- check:build, check:unit, check:lint: pass.
- show:test: shown (happy `TestExecuteUpdateQuery_Success`, unhappy `TestExecuteUpdateQuery_CommitFailureIsReported`).
- failfirst: skipped with a reason (revert mode cannot compile, the tests use seams that are part of the change;
  checked by hand, 9 new tests go red with the old discard-the-result behaviour).
- Self-review: two passes, 1 blocking (call-site wiring untested), 3 tidy, all fixed.

## Next step

Run on a worker whose acceptance runner is free: `agent-fleet gates run --owed`, then `gates run --final`,
`gates push -u origin HEAD`, open the PR titled
`fix(connection): report a failed save of connection state. Fixes #5058` with the body in the PR step
(shape "What changed / Proof", the repo's template: closing line, `Changelog:` line, What changed, Proof).

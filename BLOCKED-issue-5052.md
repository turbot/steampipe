# BLOCKED: issue 5052

## Kind: environment

The change is implemented, self-reviewed and checkpoint-pushed to `fix/issue-5052-max-parallel-bound`, but the owed `acceptance` check is unmet on this worker: its `steampipe-ports` need fails because another run's `tests/acceptance/run-local.sh` is running and port 9193 is in use. I did not touch that process. The repo's `open-incomplete` action is off, so I cannot open the PR with the check unproven.

What was done: `MaxDbConnections` returns `(int32, error)` and rejects values outside the int32 range with an error naming the max-parallel setting, `STEAMPIPE_MAX_PARALLEL`, `max_parallel` and 2147483647; `steampipe query` checks it before starting the service; the `//nolint:gosec` is removed; unit tests in `pkg/db/db_common/max_connections_test.go`; acceptance cases in `tests/acceptance/test_files/exit_codes.bats`. Build, unit and lint checks passed; show:test recorded; fail-first recorded as skipped with a reason.

Next step: when the ports are free, run `agent-fleet gates run --owed --tests 'go test ./pkg/db/db_common/ -run TestMaxDbConnections'`, then `agent-fleet gates push -u origin HEAD` and open the PR (HANDOFF.md in the wip folder has the body).

Note: the issue names a `--max-parallel` flag, but no command in this repo registers it (`steampipe query --max-parallel 1` is "unknown flag"); the env var and `max_parallel` reach the check. I added no flag.

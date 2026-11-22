package db_local

import "context"

// pgHbaCarryOverError stops an upgrade because the rule files of the older install's pg_hba.conf.d could not be
// copied. Its message says what to do, so callers pass it on as it is.
type pgHbaCarryOverError struct{ message string }

func (e *pgHbaCarryOverError) Error() string { return e.message }

// carryOverPgHbaRules does not copy anything yet.
func carryOverPgHbaRules(ctx context.Context) error { return nil }

// copyPgHbaRules does not copy anything yet.
func copyPgHbaRules(oldDataDir, newDataDir string) error { return nil }

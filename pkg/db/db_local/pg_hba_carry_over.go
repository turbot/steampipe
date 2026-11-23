package db_local

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

// pgHbaCarryOverError stops an upgrade because the rule files of the older install's pg_hba.conf.d could not be
// copied. Its message says what to do, so callers pass it on as it is.
type pgHbaCarryOverError struct{ message string }

func (e *pgHbaCarryOverError) Error() string { return e.message }

// carryOverPgHbaRules copies the rule files in the pg_hba.conf.d of the newest older install into this version's
// data directory, which an upgrade across Postgres major versions starts empty. Run it before Steampipe's
// pg_hba.conf is written there: that file includes them. Nothing is copied when there is no older install, when
// the user manages the older pg_hba.conf themselves (the files were not in use, and copying them would put
// rules in force that were not), or when the folder has no rule files.
func carryOverPgHbaRules(ctx context.Context) error {
	found, location, err := findDifferentPgInstallation(ctx)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	return copyPgHbaRules(filepath.Join(location, "data"), filepaths.GetDataLocation())
}

// checkPgHbaRulesReadable reads what carryOverPgHbaRules will copy, and stops with the same error it would give.
// The install runs it before it sets up the new data folder: a failure after initdb leaves a folder that the next
// start refuses, while a failure before it leaves nothing to clear.
func checkPgHbaRulesReadable(ctx context.Context) error {
	found, location, err := findDifferentPgInstallation(ctx)
	if err != nil || !found {
		return err
	}
	_, err = pgHbaRulesToCarry(filepath.Join(location, "data"), filepaths.GetDataLocation(), false)
	return err
}

// pgHbaRulesToCarry reads the rule files in oldDataDir's pg_hba.conf.d, or none when the user manages
// oldDataDir's pg_hba.conf themselves.
func pgHbaRulesToCarry(oldDataDir, newDataDir string, afterInit bool) ([]pgHbaRuleFile, error) {
	owner, err := pgHbaOwnerOf(oldDataDir)
	if err != nil {
		return nil, carryOverError(oldDataDir, newDataDir, err, afterInit)
	}
	if owner != pgHbaSteampipe {
		log.Printf("[TRACE] %s is not Steampipe's, so the files in %s are not copied", filepath.Join(oldDataDir, pgHbaFileName), pgHbaDirName)
		return nil, nil
	}
	rules, err := readPgHbaUserRules(filepath.Join(oldDataDir, pgHbaDirName))
	if err != nil {
		return nil, carryOverError(oldDataDir, newDataDir, err, afterInit)
	}
	return rules, nil
}

// copyPgHbaRules copies the rule files from oldDataDir's pg_hba.conf.d to newDataDir's, under the same names,
// when Steampipe wrote oldDataDir's pg_hba.conf.
func copyPgHbaRules(oldDataDir, newDataDir string) error {
	rules, err := pgHbaRulesToCarry(oldDataDir, newDataDir, true)
	if err != nil || len(rules) == 0 {
		return err
	}
	newHbaDir := filepath.Join(newDataDir, pgHbaDirName)
	if err := os.MkdirAll(newHbaDir, 0700); err != nil {
		return carryOverError(oldDataDir, newDataDir, err, true)
	}
	for _, r := range rules {
		if err := os.WriteFile(filepath.Join(newHbaDir, r.name), []byte(r.content), 0600); err != nil {
			return carryOverError(oldDataDir, newDataDir, err, true)
		}
	}
	return nil
}

// carryOverError says which copy failed. The install reads the files before it sets up the new data folder, so a
// failure then leaves nothing to clear. A failure after initdb (afterInit) finds a folder that holds a database:
// the user moves it away first, as for any stopped setup.
func carryOverError(oldDataDir, newDataDir string, err error, afterInit bool) error {
	oldHbaDir, newHbaDir := filepath.Join(oldDataDir, pgHbaDirName), filepath.Join(newDataDir, pgHbaDirName)
	fix := "Fix the problem named above, then run Steampipe again."
	if afterInit {
		fix = fmt.Sprintf("Fix the problem named above, move or delete %s, then run Steampipe again.", newDataDir)
	}
	return &pgHbaCarryOverError{message: fmt.Sprintf("could not copy your rules from %s to %s: %s\n\nYour rules are still in %s. %s",
		oldHbaDir, newHbaDir, err, oldHbaDir, fix)}
}

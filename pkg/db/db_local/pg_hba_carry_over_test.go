package db_local

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

func readRuleFileNames(t *testing.T, dataDir string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(dataDir, "pg_hba.conf.d"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestCopyPgHbaRules_CopiesTheRuleFilesAndTheNewFileUsesThem(t *testing.T) {
	oldDir, newDir := t.TempDir(), t.TempDir()
	steampipeFileWith(t, oldDir, map[string]string{
		"10-mine.conf":   "host all bob 10.0.0.0/8 reject\n",
		"20-theirs.conf": "host all alice 10.0.0.0/8 md5\n",
		"notes.txt":      "not a rule file",
		".hidden.conf":   "host all eve all trust\n",
	})

	if err := copyPgHbaRules(oldDir, newDir); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(readRuleFileNames(t, newDir), ","); got != "10-mine.conf,20-theirs.conf" {
		t.Errorf("only the files that postgres would read are copied, got %q", got)
	}
	// the data directory's pg_hba.conf is written after the copy and includes the rules, in name order
	if err := writeSteampipePgHba(newDir); err != nil {
		t.Fatal(err)
	}
	hba := readHba(t, newDir)
	bob, alice := strings.Index(hba, "host all bob 10.0.0.0/8 reject"), strings.Index(hba, "host all alice 10.0.0.0/8 md5")
	steampipe := strings.Index(hba, "# Steampipe's rules")
	if bob < 0 || alice < bob || steampipe < alice {
		t.Errorf("the user's rules must come first, in name order, before Steampipe's:\n%s", hba)
	}
	if strings.Contains(hba, "eve") {
		t.Errorf("a hidden file is not a rule file:\n%s", hba)
	}
	// the old folder is left as it was: the restore may still need it
	if got := len(readRuleFileNames(t, oldDir)); got != 4 {
		t.Errorf("the old pg_hba.conf.d must be left alone, has %d files", got)
	}
}

func TestCopyPgHbaRules_NothingToCopy(t *testing.T) {
	oldDir, newDir := t.TempDir(), t.TempDir()
	steampipeFileWith(t, oldDir, nil)
	if err := copyPgHbaRules(oldDir, newDir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(newDir, "pg_hba.conf.d")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("no rule files, so nothing is created: %v", err)
	}
	// no folder and no pg_hba.conf at all: an install with no rules of the user's
	if err := copyPgHbaRules(t.TempDir(), newDir); err != nil {
		t.Errorf("an old data directory with no rules is not an error: %v", err)
	}
}

func TestCopyPgHbaRules_DoesNotCopyFilesThatTheOlderInstallNeverUsed(t *testing.T) {
	// a pg_hba.conf that Steampipe did not write is the user's own, and Steampipe never read the folder; copying
	// would put rules in force that were not
	oldDir, newDir := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(oldDir, "pg_hba.conf"), "local all all peer\n")
	writeFile(t, filepath.Join(oldDir, "pg_hba.conf.d", "10-mine.conf"), "host all bob all trust\n")
	if err := copyPgHbaRules(oldDir, newDir); err != nil {
		t.Fatal(err)
	}
	if names := readRuleFileNames(t, newDir); len(names) != 0 {
		t.Errorf("rules that were not in use must not be copied, got %v", names)
	}
}

func TestCopyPgHbaRules_FailsWithAMessageThatNamesTheFolders(t *testing.T) {
	oldDir, newDir := t.TempDir(), t.TempDir()
	steampipeFileWith(t, oldDir, map[string]string{"10-mine.conf": "host all bob all reject\n"})
	// a file where the new folder must go
	writeFile(t, filepath.Join(newDir, "pg_hba.conf.d"), "")

	err := copyPgHbaRules(oldDir, newDir)
	var carry *pgHbaCarryOverError
	if !errors.As(err, &carry) {
		t.Fatalf("a rule that could not be copied must stop the upgrade, got %v", err)
	}
	for _, want := range []string{filepath.Join(oldDir, "pg_hba.conf.d"), filepath.Join(newDir, "pg_hba.conf.d")} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the message must name %s:\n%s", want, err)
		}
	}
}

func TestCarryOverPgHbaRules_UsesTheNewestOlderInstall(t *testing.T) {
	installDir := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = installDir })
	app_specific.InstallDir = filepath.Join(shortTempDir(t), ".steampipe")
	for version, rule := range map[string]string{"12.1.0": "host all old all reject\n", "13.2.0": "host all newest all reject\n"} {
		dir := filepath.Join(filepaths.EnsureDatabaseDir(), version)
		writeFile(t, filepath.Join(dir, "postgres", "bin", "postgres"), "")
		writeFile(t, filepath.Join(dir, "data", "PG_VERSION"), "13\n")
		writeFile(t, filepath.Join(dir, "data", "pg_hba.conf.d", "10-mine.conf"), rule)
	}

	if err := carryOverPgHbaRules(context.Background()); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(filepaths.GetDataLocation(), "pg_hba.conf.d", "10-mine.conf"))
	if err != nil || !strings.Contains(string(got), "newest") {
		t.Errorf("the rules come from the newest older install, got %q, %v", got, err)
	}
}

func TestCarryOverPgHbaRules_NoOlderInstall(t *testing.T) {
	installDir := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = installDir })
	app_specific.InstallDir = filepath.Join(shortTempDir(t), ".steampipe")

	if err := carryOverPgHbaRules(context.Background()); err != nil {
		t.Fatal(err)
	}
	if names := readRuleFileNames(t, filepaths.GetDataLocation()); len(names) != 0 {
		t.Errorf("a new install has no rules to carry over, got %v", names)
	}
}

func TestRunInstall_StopsBeforeTheDataFolderIsSetUpWhenTheRulesCannotBeRead(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root reads a folder whatever its mode")
	}
	installDir := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = installDir })
	app_specific.InstallDir = filepath.Join(shortTempDir(t), ".steampipe")
	older := filepath.Join(filepaths.EnsureDatabaseDir(), "12.1.0")
	writeFile(t, filepath.Join(older, "postgres", "bin", "postgres"), "")
	writeFile(t, filepath.Join(older, "data", "PG_VERSION"), "12\n")
	writeFile(t, filepath.Join(older, "data", "pg_hba.conf.d", "10-mine.conf"), "host all bob all reject\n")
	hbaDir := filepath.Join(older, "data", "pg_hba.conf.d")
	if err := os.Chmod(hbaDir, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(hbaDir, 0700) })

	err := runInstall(context.Background(), nil)
	var carry *pgHbaCarryOverError
	if !errors.As(err, &carry) {
		t.Fatalf("an unreadable rule folder must stop the install with its own message, got %v", err)
	}
	if !strings.Contains(err.Error(), "then run Steampipe again") || strings.Contains(err.Error(), "move or delete") {
		t.Errorf("before the new data folder exists the only step is to fix the folder:\n%s", err)
	}
	// nothing was set up, so the next run is not refused by a half-made data folder
	if _, statErr := os.Stat(filepaths.GetPgVersionLocation()); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("the new data folder must not hold a database: %v", statErr)
	}
	if names := readRuleFileNames(t, filepaths.GetDataLocation()); len(names) != 0 {
		t.Errorf("nothing is copied before the check passes, got %v", names)
	}
}

func TestCopyPgHbaRules_AWriteFailureTellsTheUserToMoveTheNewFolder(t *testing.T) {
	oldDir, newDir := t.TempDir(), t.TempDir()
	steampipeFileWith(t, oldDir, map[string]string{"10-mine.conf": "host all bob all reject\n"})
	writeFile(t, filepath.Join(newDir, "pg_hba.conf.d"), "")

	err := copyPgHbaRules(oldDir, newDir)
	if err == nil || !strings.Contains(err.Error(), "move or delete "+newDir) {
		t.Errorf("after initdb the new folder holds a database, so the message must say to move it:\n%v", err)
	}
}

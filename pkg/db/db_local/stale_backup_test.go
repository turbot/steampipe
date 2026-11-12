package db_local

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

func stubInstallSteps(t *testing.T) {
	t.Helper()
	origDownload, origFdw, origInit := installDbFiles, installFdwFiles, initNewDatabase
	t.Cleanup(func() { installDbFiles, installFdwFiles, initNewDatabase = origDownload, origFdw, origInit })
	installDbFiles = func(context.Context) error { return nil }
	installFdwFiles = func(context.Context, bool) (string, error) { return "", nil }
	initNewDatabase = func(context.Context, *string) error {
		t.Error("the database must not be initialised again")
		return nil
	}
}

func writeCurrentVersionDatabase(t *testing.T) {
	t.Helper()
	if err := os.WriteFile(filepaths.GetPgVersionLocation(), []byte("14\n"), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureDBInstalled_KeptDatabaseDoesNotGetALeftoverBackupRestored(t *testing.T) {
	setCleanupInstallDir(t)
	stubInstallSteps(t)
	writeCurrentVersionDatabase(t)

	backupFile := filepaths.DatabaseBackupFilePath()
	if err := os.WriteFile(backupFile, []byte("dump of an old database"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := EnsureDBInstalled(context.Background()); err != nil {
		t.Fatal(err)
	}

	assertGone(t, backupFile)
	kept, err := filepath.Glob(filepath.Join(filepaths.EnsureBackupsDir(), "database-*.dump"))
	if err != nil || len(kept) != 1 {
		t.Fatalf("expected the backup to be kept in the backups folder, got %v, %v", kept, err)
	}
	if content, err := os.ReadFile(kept[0]); err != nil || string(content) != "dump of an old database" {
		t.Errorf("kept backup content changed: %q, %v", content, err)
	}
	assertExists(t, filepaths.GetPgVersionLocation())
}

func TestEnsureDBInstalled_BackupThatCannotBeSetAsideRemovesThePrograms(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("a read-only folder does not stop root")
	}
	setCleanupInstallDir(t)
	stubInstallSteps(t)
	writeCurrentVersionDatabase(t)
	programs := filepath.Join(filepaths.GetDatabaseLocation(), "programs")
	installDbFiles = func(context.Context) error {
		return os.WriteFile(programs, []byte(""), 0600)
	}

	backupFile := filepaths.DatabaseBackupFilePath()
	if err := os.WriteFile(backupFile, []byte("dump of an old database"), 0600); err != nil {
		t.Fatal(err)
	}
	backupsDir := filepaths.EnsureBackupsDir()
	if err := os.Chmod(backupsDir, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(backupsDir, 0700) })

	if err := EnsureDBInstalled(context.Background()); err == nil {
		t.Fatal("expected an error when the backup cannot be moved out of the way")
	}

	assertExists(t, backupFile)
	assertGone(t, programs)
	assertExists(t, filepaths.GetPgVersionLocation())
}

func TestEnsureDBInstalled_KeptDatabaseWithNoBackupKeepsNothing(t *testing.T) {
	setCleanupInstallDir(t)
	stubInstallSteps(t)
	writeCurrentVersionDatabase(t)

	if err := EnsureDBInstalled(context.Background()); err != nil {
		t.Fatal(err)
	}

	if kept, _ := filepath.Glob(filepath.Join(filepaths.EnsureBackupsDir(), "*")); len(kept) != 0 {
		t.Errorf("nothing should be kept when there was no backup: %v", kept)
	}
}

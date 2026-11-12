package db_local

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

// stubDownloads replaces the download and FDW install of EnsureDBInstalled with steps that succeed and install nothing
func stubDownloads(t *testing.T) {
	t.Helper()
	origDownload, origFdw := installDbFiles, installFdwFiles
	t.Cleanup(func() { installDbFiles, installFdwFiles = origDownload, origFdw })
	installDbFiles = func(context.Context) error { return nil }
	installFdwFiles = func(context.Context, bool) (string, error) { return "", nil }
}

func writeCurrentVersionData(t *testing.T) string {
	t.Helper()
	pgVersion := filepaths.GetPgVersionLocation()
	if err := os.MkdirAll(filepath.Dir(pgVersion), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pgVersion, []byte("14\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return pgVersion
}

// the current version's programs are missing but its data folder holds a database, and no other version is
// installed: the database is used as it is, so initialising a new one, which would fail here, is never attempted
func TestEnsureDBInstalled_ReusesCurrentVersionDatabaseWhenOnlyProgramsAreMissing(t *testing.T) {
	setCleanupInstallDir(t)
	stubDownloads(t)
	pgVersion := writeCurrentVersionData(t)

	origInit := initNewDatabase
	t.Cleanup(func() { initNewDatabase = origInit })
	initNewDatabase = func(context.Context, *string) error {
		return errors.New("the existing database must not be initialised again")
	}

	if err := EnsureDBInstalled(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertExists(t, pgVersion)
}

// an older folder holding programs but no database is not another version's database, so it does not stop the
// current version's database being reused
func TestEnsureDBInstalled_ReusesCurrentVersionDatabaseWithAnOlderFolderHoldingNoDatabase(t *testing.T) {
	dbDir := setCleanupInstallDir(t)
	stubDownloads(t)
	pgVersion := writeCurrentVersionData(t)
	writeProgramsOnlyInstall(t, filepath.Join(dbDir, "13.0.0"))

	origInit := initNewDatabase
	t.Cleanup(func() { initNewDatabase = origInit })
	initNewDatabase = func(context.Context, *string) error {
		return errors.New("the existing database must not be initialised again")
	}

	if err := EnsureDBInstalled(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertExists(t, pgVersion)
}

// another version's database is installed, and the current version's data folder already holds a database: the
// install is refused and neither database is emptied
func TestEnsureDBInstalled_NeverEmptiesADatabaseWhenAnotherVersionIsInstalled(t *testing.T) {
	dbDir := setCleanupInstallDir(t)
	stubDownloads(t)
	pgVersion := writeCurrentVersionData(t)

	target, err := semver.NewVersion(constants.DatabaseVersion)
	if err != nil {
		t.Fatal(err)
	}
	// a different major version, so its data is not moved into the current version's data folder
	other := filepath.Join(dbDir, fmt.Sprintf("%d.0.0", target.Major()-1))
	writeFakeInstall(t, other)

	err = EnsureDBInstalled(context.Background())
	if !errors.Is(err, errDataFolderHoldsDatabase) {
		t.Fatalf("expected errDataFolderHoldsDatabase, got: %v", err)
	}
	assertExists(t, pgVersion)
	assertExists(t, filepath.Join(other, "data", "PG_VERSION"))
}

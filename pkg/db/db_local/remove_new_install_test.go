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

func TestRemoveNewInstall_KeepsNonEmptyData(t *testing.T) {
	original := app_specific.InstallDir
	defer func() { app_specific.InstallDir = original }()
	app_specific.InstallDir = filepath.Join(t.TempDir(), ".steampipe")

	newLocation := filepaths.DatabaseInstanceDir()
	newBinaries := filepaths.GetDatabaseLocation()
	if err := os.WriteFile(filepath.Join(newBinaries, "marker"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}
	newDataLocation := filepaths.GetDataLocation()
	if err := os.WriteFile(filepath.Join(newDataLocation, "PG_VERSION"), []byte("14\n"), 0600); err != nil {
		t.Fatal(err)
	}

	if !removeNewInstall(newLocation, newDataLocation) {
		t.Error("expected the non-empty data folder to be reported as kept")
	}
	if _, err := os.Stat(filepath.Join(newDataLocation, "PG_VERSION")); err != nil {
		t.Errorf("new data should have been kept: %v", err)
	}
	if _, err := os.Stat(newBinaries); !os.IsNotExist(err) {
		t.Errorf("new binaries %s should have been removed, stat err: %v", newBinaries, err)
	}
}

func TestAbortInstallAfterRunningInstance_KeepsNonEmptyData(t *testing.T) {
	original := app_specific.InstallDir
	defer func() { app_specific.InstallDir = original }()
	app_specific.InstallDir = filepath.Join(t.TempDir(), ".steampipe")

	newBinaries := filepaths.GetDatabaseLocation()
	if err := os.WriteFile(filepath.Join(newBinaries, "marker"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}
	newDataLocation := filepaths.GetDataLocation()
	if err := os.WriteFile(filepath.Join(newDataLocation, "PG_VERSION"), []byte("14\n"), 0600); err != nil {
		t.Fatal(err)
	}

	err := abortInstallRemovingNewInstall(errDbInstanceRunning, "the old server is stopped")
	if !errors.Is(err, errDbInstanceRunning) {
		t.Errorf("error should wrap errDbInstanceRunning: %v", err)
	}
	if !strings.Contains(err.Error(), newDataLocation) {
		t.Errorf("error should name the kept data folder %s: %v", newDataLocation, err)
	}
	if _, statErr := os.Stat(filepath.Join(newDataLocation, "PG_VERSION")); statErr != nil {
		t.Errorf("new data should have been kept: %v", statErr)
	}
	if _, statErr := os.Stat(newBinaries); !os.IsNotExist(statErr) {
		t.Errorf("new binaries %s should have been removed, stat err: %v", newBinaries, statErr)
	}
}

func TestAbortInstallAfterRunningInstance_EmptyDataReturnsErrorUnchanged(t *testing.T) {
	original := app_specific.InstallDir
	defer func() { app_specific.InstallDir = original }()
	app_specific.InstallDir = filepath.Join(t.TempDir(), ".steampipe")

	newLocation := filepaths.DatabaseInstanceDir()
	_ = filepaths.GetDatabaseLocation()
	_ = filepaths.GetDataLocation()

	if err := abortInstallRemovingNewInstall(errDbInstanceRunning, "the old server is stopped"); err != errDbInstanceRunning {
		t.Errorf("error should be returned unchanged: %v", err)
	}
	if _, statErr := os.Stat(newLocation); !os.IsNotExist(statErr) {
		t.Errorf("new installation should have been removed, stat err: %v", statErr)
	}
}

func TestRemoveNewInstall_RemovesEmptyData(t *testing.T) {
	original := app_specific.InstallDir
	defer func() { app_specific.InstallDir = original }()
	app_specific.InstallDir = filepath.Join(t.TempDir(), ".steampipe")

	newLocation := filepaths.DatabaseInstanceDir()
	_ = filepaths.GetDatabaseLocation()
	newDataLocation := filepaths.GetDataLocation()

	if removeNewInstall(newLocation, newDataLocation) {
		t.Error("an empty data folder should not be reported as kept")
	}
	if _, err := os.Stat(newLocation); !os.IsNotExist(err) {
		t.Errorf("new installation %s should have been removed, stat err: %v", newLocation, err)
	}
}

func TestPrepareDataLocation_RefusesToEmptyADatabase(t *testing.T) {
	dataLocation := t.TempDir()
	if err := os.WriteFile(filepath.Join(dataLocation, "PG_VERSION"), []byte("14\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataLocation, "postgresql.conf"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}

	err := prepareDataLocation(dataLocation)
	if !errors.Is(err, errDataFolderHoldsDatabase) {
		t.Fatalf("expected errDataFolderHoldsDatabase, got: %v", err)
	}
	if !strings.Contains(err.Error(), dataLocation) {
		t.Errorf("error should name %s: %v", dataLocation, err)
	}
	for _, kept := range []string{"PG_VERSION", "postgresql.conf"} {
		if _, statErr := os.Stat(filepath.Join(dataLocation, kept)); statErr != nil {
			t.Errorf("%s should have been kept: %v", kept, statErr)
		}
	}
}

func TestRunInstall_ReturnsTheRefusalToEmptyADatabase(t *testing.T) {
	original := app_specific.InstallDir
	defer func() { app_specific.InstallDir = original }()
	app_specific.InstallDir = filepath.Join(shortTempDir(t), ".steampipe")

	newDataLocation := filepaths.GetDataLocation()
	if err := os.WriteFile(filepath.Join(newDataLocation, "PG_VERSION"), []byte("14\n"), 0600); err != nil {
		t.Fatal(err)
	}

	err := runInstall(context.Background(), nil)
	if !errors.Is(err, errDataFolderHoldsDatabase) {
		t.Fatalf("expected errDataFolderHoldsDatabase, got: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(newDataLocation, "PG_VERSION")); statErr != nil {
		t.Errorf("PG_VERSION should have been kept: %v", statErr)
	}
}

func TestPrepareDataLocation_EmptiesAFolderWithoutADatabase(t *testing.T) {
	dataLocation := t.TempDir()
	// a failed initdb leaves files behind but no PG_VERSION
	if err := os.WriteFile(filepath.Join(dataLocation, "postgresql.conf"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}

	if err := prepareDataLocation(dataLocation); err != nil {
		t.Fatal(err)
	}
	if entries, err := os.ReadDir(dataLocation); err != nil || len(entries) != 0 {
		t.Errorf("data folder should be empty: %v, %v", entries, err)
	}
}

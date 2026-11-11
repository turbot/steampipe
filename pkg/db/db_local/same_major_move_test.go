package db_local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

func TestIsSameMajorVersion(t *testing.T) {
	tests := []struct {
		name, old, target string
		want              bool
	}{
		{"same major, minor bump", "14.19.0", "14.24.0", true},
		{"same major, older minor", "14.17.0", "14.19.0", true},
		{"major upgrade", "14.19.0", "18.1.0", false},
		{"major downgrade", "18.1.0", "14.19.0", false},
		{"unparseable old version", "not-a-version", "14.24.0", false},
		{"empty target version", "14.19.0", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSameMajorVersion(tt.old, tt.target); got != tt.want {
				t.Errorf("isSameMajorVersion(%q, %q) = %v, want %v", tt.old, tt.target, got, tt.want)
			}
		})
	}
}

// writeFakeInstall creates a version folder holding a postgres binary and a data folder with one marker file
func writeFakeInstall(t *testing.T, location string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(location, "postgres", "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(location, "postgres", "bin", "postgres"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(location, "data"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(location, "data", "PG_VERSION"), []byte("14\n"), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestFindDifferentPgInstallation_SameMajor(t *testing.T) {
	original := app_specific.InstallDir
	defer func() { app_specific.InstallDir = original }()
	app_specific.InstallDir = filepath.Join(t.TempDir(), ".steampipe")

	target, err := semver.NewVersion(constants.DatabaseVersion)
	if err != nil {
		t.Fatal(err)
	}
	oldLocation := filepath.Join(app_specific.InstallDir, "db", fmt.Sprintf("%d.0.0", target.Major()))
	writeFakeInstall(t, oldLocation)

	found, location, err := findDifferentPgInstallation(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !found || location != oldLocation {
		t.Fatalf("findDifferentPgInstallation = %v, %q; want true, %q", found, location, oldLocation)
	}
	if !isSameMajorVersion(filepath.Base(location), constants.DatabaseVersion) {
		t.Fatalf("%s should be the same major as %s", filepath.Base(location), constants.DatabaseVersion)
	}
}

func TestMoveDataFolder(t *testing.T) {
	dbDir := t.TempDir()
	oldLocation := filepath.Join(dbDir, "14.19.0")
	newLocation := filepath.Join(dbDir, "14.24.0")
	writeFakeInstall(t, oldLocation)
	newDataLocation := filepath.Join(newLocation, "data")
	// the path helpers create the new data folder empty before the move
	if err := os.MkdirAll(newDataLocation, 0755); err != nil {
		t.Fatal(err)
	}

	if err := moveDataFolder(oldLocation, newDataLocation); err != nil {
		t.Fatalf("moveDataFolder failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(newDataLocation, "PG_VERSION"))
	if err != nil || string(content) != "14\n" {
		t.Fatalf("moved data folder does not hold the old data: %q, %v", content, err)
	}
	info, err := os.Stat(newDataLocation)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0700 {
		t.Errorf("moved data folder has mode %v, want 0700", info.Mode().Perm())
	}
	if _, err := os.Stat(oldLocation); !os.IsNotExist(err) {
		t.Errorf("old installation %s should have been removed, stat err: %v", oldLocation, err)
	}
}

func TestMoveDataFromSameMajorInstall_FailureRemovesNewInstall(t *testing.T) {
	original := app_specific.InstallDir
	defer func() { app_specific.InstallDir = original }()
	app_specific.InstallDir = filepath.Join(t.TempDir(), ".steampipe")

	// the fake postgres binary cannot be executed, so starting the old server fails
	oldLocation := filepath.Join(app_specific.InstallDir, "db", "14.19.0")
	writeFakeInstall(t, oldLocation)
	newLocation := filepaths.DatabaseInstanceDir()

	err := moveDataFromSameMajorInstall(context.Background(), oldLocation, filepaths.GetDataLocation())
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), oldLocation) || !strings.Contains(err.Error(), newLocation) {
		t.Errorf("error should name %s and %s: %v", oldLocation, newLocation, err)
	}
	content, readErr := os.ReadFile(filepath.Join(oldLocation, "data", "PG_VERSION"))
	if readErr != nil || string(content) != "14\n" {
		t.Fatalf("old data folder was changed: %q, %v", content, readErr)
	}
	if _, statErr := os.Stat(newLocation); !os.IsNotExist(statErr) {
		t.Errorf("new installation %s should have been removed, stat err: %v", newLocation, statErr)
	}
}

func TestMoveDataFolder_FailureLeavesOldInstall(t *testing.T) {
	dbDir := t.TempDir()
	oldLocation := filepath.Join(dbDir, "14.19.0")
	writeFakeInstall(t, oldLocation)

	tests := map[string]func(t *testing.T) string{
		// the rename itself fails
		"missing parent": func(t *testing.T) string {
			return filepath.Join(dbDir, "missing", "data")
		},
		// the new data folder is not empty, so it is never replaced
		"non-empty target": func(t *testing.T) string {
			target := filepath.Join(dbDir, "14.24.0", "data")
			if err := os.MkdirAll(target, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(target, "other"), []byte(""), 0600); err != nil {
				t.Fatal(err)
			}
			return target
		},
	}
	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			if err := moveDataFolder(oldLocation, setup(t)); err == nil {
				t.Fatal("expected an error")
			}
			content, err := os.ReadFile(filepath.Join(oldLocation, "data", "PG_VERSION"))
			if err != nil || string(content) != "14\n" {
				t.Fatalf("old data folder was changed: %q, %v", content, err)
			}
			if _, err := os.Stat(filepath.Join(oldLocation, "postgres", "bin", "postgres")); err != nil {
				t.Fatalf("old binaries were changed: %v", err)
			}
		})
	}
}

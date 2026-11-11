package db_local

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

func TestIsSameMajorVersion(t *testing.T) {
	tests := []struct {
		old, target string
		want        bool
	}{
		{"14.19.0", "14.24.0", true},
		{"14.17.0", "14.19.0", true},
		{"14.19.0", "18.1.0", false},
		{"18.1.0", "14.19.0", false},
		{"not-a-version", "14.24.0", false},
		{"14.19.0", "", false},
	}
	for _, tt := range tests {
		if got := isSameMajorVersion(tt.old, tt.target); got != tt.want {
			t.Errorf("isSameMajorVersion(%q, %q) = %v, want %v", tt.old, tt.target, got, tt.want)
		}
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

func TestFindDifferentPgInstallationSameMajor(t *testing.T) {
	original := app_specific.InstallDir
	defer func() { app_specific.InstallDir = original }()
	app_specific.InstallDir = filepath.Join(t.TempDir(), ".steampipe")

	oldLocation := filepath.Join(app_specific.InstallDir, "db", "14.0.0")
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

func TestMoveDataFolderFailureLeavesOldInstall(t *testing.T) {
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

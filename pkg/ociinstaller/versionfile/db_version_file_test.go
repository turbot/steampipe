package versionfile

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteDatabaseVersionFile(t *testing.T) {

	var v DatabaseVersionFile

	fileName := "test.json"
	timeNow := time.Now()

	v.EmbeddedDB.Version = "0.0.1"
	v.EmbeddedDB.Name = "embeddedDb"
	v.EmbeddedDB.ImageDigest = "111111111111"
	v.EmbeddedDB.InstalledFrom = "hub.steampipe.io/core/embedded-postgres:latest"
	v.EmbeddedDB.LastCheckedDate = timeNow.Format(time.UnixDate)
	v.EmbeddedDB.InstallDate = timeNow.Format(time.UnixDate)

	timeNow2 := timeNow.Add(time.Minute * 10)

	v.FdwExtension.Version = "1.0.1"
	v.FdwExtension.Name = "fdwExtension"
	v.FdwExtension.ImageDigest = "2222222222"
	v.FdwExtension.InstalledFrom = "hub.steampipe.io/core/hub-extension:latest"
	v.FdwExtension.LastCheckedDate = timeNow2.Format(time.UnixDate)
	v.FdwExtension.InstallDate = timeNow2.Format(time.UnixDate)

	if err := v.write(fileName); err != nil {
		t.Errorf("\nError writing file: %s", err.Error())
	}

	v2, err := readDatabaseVersionFile(fileName)
	if err != nil {
		t.Errorf("\nError reading file: %s", err.Error())
	}

	if v2.EmbeddedDB.Version != v.EmbeddedDB.Version {
		t.Errorf("\nError EmbeddedDB.Version is: %s, expected %s", v2.EmbeddedDB.Version, v.EmbeddedDB.Version)
	}
	if v2.EmbeddedDB.Name != v.EmbeddedDB.Name {
		t.Errorf("\nError EmbeddedDB.Name is: %s, expected %s", v2.EmbeddedDB.Name, v.EmbeddedDB.Name)
	}
	if v2.EmbeddedDB.ImageDigest != v.EmbeddedDB.ImageDigest {
		t.Errorf("\nError EmbeddedDB.ImageDigest is: %s, expected %s", v2.EmbeddedDB.ImageDigest, v.EmbeddedDB.ImageDigest)
	}
	if v2.EmbeddedDB.InstalledFrom != v.EmbeddedDB.InstalledFrom {
		t.Errorf("\nError EmbeddedDB.InstalledFrom is: %s, expected %s", v2.EmbeddedDB.InstalledFrom, v.EmbeddedDB.InstalledFrom)
	}
	if v2.EmbeddedDB.LastCheckedDate != v.EmbeddedDB.LastCheckedDate {
		t.Errorf("\nError EmbeddedDB.LastCheckedDate is: %s, expected %s", v2.EmbeddedDB.LastCheckedDate, v.EmbeddedDB.LastCheckedDate)
	}
	if v2.EmbeddedDB.InstallDate != v.EmbeddedDB.InstallDate {
		t.Errorf("\nError EmbeddedDB.InstallDate is: %s, expected %s", v2.EmbeddedDB.InstallDate, v.EmbeddedDB.InstallDate)
	}

	os.Remove(fileName)
}

func TestWriteDatabaseVersionFilePermissions(t *testing.T) {
	var v DatabaseVersionFile

	path := filepath.Join(t.TempDir(), "test.json")

	if err := v.write(path); err != nil {
		t.Fatalf("Error writing file: %s", err.Error())
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat %s: %v", path, err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected %s to have permissions 0600, got %o", path, perm)
	}
}

package steampipeconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
)

func setConfigExtension(t *testing.T) {
	t.Helper()
	prev := app_specific.ConfigExtension
	app_specific.ConfigExtension = ".spc"
	t.Cleanup(func() { app_specific.ConfigExtension = prev })
}

func TestValidConfigDir_AllFilesValidReturnsSameDir(t *testing.T) {
	setConfigExtension(t)
	dir := t.TempDir()
	writeConfigFile(t, dir, "a.spc", connectionConfig("a"))

	got, warnings, cleanup, err := ValidConfigDir(dir)
	defer cleanup()
	if err != nil {
		t.Fatal(err)
	}
	if got != dir || len(warnings) != 0 {
		t.Errorf("want same dir and no warnings, got %q, %v", got, warnings)
	}
}

func TestValidConfigDir_EmptyDirReturnsSameDir(t *testing.T) {
	setConfigExtension(t)
	dir := t.TempDir()

	got, warnings, cleanup, err := ValidConfigDir(dir)
	defer cleanup()
	if err != nil || got != dir || len(warnings) != 0 {
		t.Errorf("want same dir, got %q, %v, %v", got, warnings, err)
	}
}

func TestValidConfigDir_BadFileIsLeftOutAndNamed(t *testing.T) {
	setConfigExtension(t)
	dir := t.TempDir()
	writeConfigFile(t, dir, "good.spc", connectionConfig("good"))
	writeConfigFile(t, dir, "broken.spc", "connection \"x\" {\n")
	writeConfigFile(t, dir, "notes.txt", "not config")

	got, warnings, cleanup, err := ValidConfigDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got == dir {
		t.Fatal("expected a copy of the folder without the bad file")
	}
	entries, err := os.ReadDir(got)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "good.spc" {
		t.Errorf("want only good.spc in the copy, got %v", entries)
	}
	content, _ := os.ReadFile(filepath.Join(got, "good.spc"))
	if string(content) != connectionConfig("good") {
		t.Errorf("good.spc content changed: %q", content)
	}
	if len(warnings) == 0 || !strings.Contains(strings.Join(warnings, "\n"), "broken.spc") {
		t.Errorf("want a warning naming broken.spc, got %v", warnings)
	}
	// the original folder is untouched
	if _, err := os.Stat(filepath.Join(dir, "broken.spc")); err != nil {
		t.Errorf("broken.spc should remain in the original folder: %v", err)
	}

	cleanup()
	if _, err := os.Stat(got); !os.IsNotExist(err) {
		t.Errorf("cleanup should remove the copy, stat err: %v", err)
	}
}

func TestValidConfigDir_OnlyBadFilesGivesEmptyCopy(t *testing.T) {
	setConfigExtension(t)
	dir := t.TempDir()
	writeConfigFile(t, dir, "broken.spc", "connection \"x\" {\n")

	got, warnings, cleanup, err := ValidConfigDir(dir)
	defer cleanup()
	if err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(got)
	if got == dir || len(entries) != 0 || len(warnings) == 0 {
		t.Errorf("want an empty copy and a warning, got %q, %v, %v", got, entries, warnings)
	}
}

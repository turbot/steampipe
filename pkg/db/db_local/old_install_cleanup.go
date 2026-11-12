package db_local

import (
	"log"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

// removeOldInstallsWithoutDatabase removes the version folders under the db folder that are older than the
// installed version and hold no database.
//
// A folder is removed only when it is a real folder named as a version older than constants.DatabaseVersion, and
// holds nothing but a "postgres" folder and a "data" folder that is absent or an empty real folder. Anything else
// is left alone, including every folder that may hold a database.
func removeOldInstallsWithoutDatabase() {
	target, err := semver.NewVersion(constants.DatabaseVersion)
	if err != nil {
		return
	}
	dbBaseDirectory := filepaths.EnsureDatabaseDir()
	entries, err := os.ReadDir(dbBaseDirectory)
	if err != nil {
		log.Printf("[WARN] Could not read %s to remove old installations: %s", dbBaseDirectory, err)
		return
	}
	for _, de := range entries {
		if !de.IsDir() {
			continue
		}
		version, err := semver.NewVersion(de.Name())
		if err != nil || !version.LessThan(target) {
			continue
		}
		location := filepath.Join(dbBaseDirectory, de.Name())
		if !holdsNoDatabase(location) {
			continue
		}
		if err := os.RemoveAll(location); err != nil {
			log.Printf("[WARN] Could not remove old installation at %s: %s", location, err)
		}
	}
}

// holdsNoDatabase reports whether a version folder holds only database programs and no data. Any doubt reports false.
func holdsNoDatabase(location string) bool {
	entries, err := os.ReadDir(location)
	if err != nil {
		return false
	}
	for _, e := range entries {
		switch e.Name() {
		case "postgres":
			if !e.IsDir() {
				return false
			}
		case "data":
			if !e.IsDir() {
				return false
			}
			dataEntries, err := os.ReadDir(filepath.Join(location, "data"))
			if err != nil || len(dataEntries) > 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

package db

import (
	"os"
	"path/filepath"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/platform"
	"github.com/turbot/steampipe/utils"
)

func databaseInstanceDir() string {
	loc := filepath.Join(constants.DatabaseDir(), constants.DatabaseVersion)
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		err = os.MkdirAll(loc, 0755)
		utils.FailOnErrorWithMessage(err, "could not ensure db version directory")
	}
	return loc
}

func getDatabaseLocation() string {
	loc := filepath.Join(databaseInstanceDir(), "postgres")
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		err = os.MkdirAll(loc, 0755)
		utils.FailOnErrorWithMessage(err, "could not ensure postgres installation directory")
	}
	return loc
}

func getDatabaseLogDirectory() string {
	loc := filepath.Join(databaseInstanceDir(), "logs")
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		err = os.MkdirAll(loc, 0755)
		utils.FailOnErrorWithMessage(err, "could not ensure postgres logging directory")
	}
	return loc
}

func getDataLocation() string {
	loc := filepath.Join(databaseInstanceDir(), "data")
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		err = os.MkdirAll(loc, 0755)
		utils.FailOnErrorWithMessage(err, "could not ensure data directory")
	}
	return loc
}

func getInitDbBinaryExecutablePath() string {
	return filepath.Join(getDatabaseLocation(), "bin", platform.Paths.InitDbExecutable)
}

func getPostgresBinaryExecutablePath() string {
	return filepath.Join(getDatabaseLocation(), "bin", platform.Paths.PostgresExecutable)
}

func getDBSignatureLocation() string {
	loc := filepath.Join(getDatabaseLocation(), "signature")
	return loc
}

func getFDWBinaryLocation() string {
	return filepath.Join(getDatabaseLocation(), "lib", "postgresql", "steampipe_postgres_fdw.so")
}

func getFDWSQLAndControlLocation() (string, string) {
	base := filepath.Join(getDatabaseLocation(), "share", "postgresql", "extension")
	sqlLocation := filepath.Join(base, "steampipe_postgres_fdw--1.0.sql")
	controlLocation := filepath.Join(base, "steampipe_postgres_fdw.control")
	return sqlLocation, controlLocation
}

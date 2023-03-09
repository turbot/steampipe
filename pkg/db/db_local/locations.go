package db_local

import (
	"os"
	"path/filepath"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/platform"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
)

func ServiceExecutableRelativeLocation() string {
	return filepath.Join("db", constants.DatabaseVersion, "postgres", "bin", "postgres")
}

func databaseInstanceDir() string {
	loc := filepath.Join(filepaths.EnsureDatabaseDir(), constants.DatabaseVersion)
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		err = os.MkdirAll(loc, 0755)
		error_helpers.FailOnErrorWithMessage(err, "could not create db version directory")
	}
	return loc
}

func getDatabaseLocation() string {
	loc := filepath.Join(databaseInstanceDir(), "postgres")
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		err = os.MkdirAll(loc, 0755)
		error_helpers.FailOnErrorWithMessage(err, "could not create postgres installation directory")
	}
	return loc
}

func getDatabaseLogDirectory() string {
	loc := filepaths.EnsureLogDir()
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		err = os.MkdirAll(loc, 0755)
		error_helpers.FailOnErrorWithMessage(err, "could not create postgres logging directory")
	}
	return loc
}

func getDataLocation() string {
	loc := filepath.Join(databaseInstanceDir(), "data")
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		err = os.MkdirAll(loc, 0755)
		error_helpers.FailOnErrorWithMessage(err, "could not create data directory")
	}
	return loc
}

// tar file where the dump file will be stored, so that it can be later restored after connections
// refresh in a new installation
func databaseBackupFilePath() string {
	return filepath.Join(filepaths.EnsureDatabaseDir(), "backup.bk")
}

func getRootCertLocation() string {
	return filepath.Join(getDataLocation(), constants.RootCert)
}

func getRootCertKeyLocation() string {
	return filepath.Join(getDataLocation(), constants.RootCertKey)
}

func getServerCertLocation() string {
	return filepath.Join(getDataLocation(), constants.ServerCert)
}

func getServerCertKeyLocation() string {
	return filepath.Join(getDataLocation(), constants.ServerCertKey)
}

func getInitDbBinaryExecutablePath() string {
	return filepath.Join(getDatabaseLocation(), "bin", platform.Paths.InitDbExecutable)
}

func getPostgresBinaryExecutablePath() string {
	return filepath.Join(getDatabaseLocation(), "bin", platform.Paths.PostgresExecutable)
}

func pgDumpBinaryExecutablePath() string {
	return filepath.Join(getDatabaseLocation(), "bin", platform.Paths.PgDumpExecutable)
}

func pgRestoreBinaryExecutablePath() string {
	return filepath.Join(getDatabaseLocation(), "bin", platform.Paths.PgRestoreExecutable)
}

func getDBSignatureLocation() string {
	loc := filepath.Join(getDatabaseLocation(), "signature")
	return loc
}

func getDatabaseLibDirectory() string {
	return filepath.Join(getDatabaseLocation(), "lib")
}

func getFDWBinaryLocation() string {
	return filepath.Join(getDatabaseLibDirectory(), "postgresql", "steampipe_postgres_fdw.so")
}

func getFDWSQLAndControlLocation() (string, string) {
	base := filepath.Join(getDatabaseLocation(), "share", "postgresql", "extension")
	sqlLocation := filepath.Join(base, "steampipe_postgres_fdw--1.0.sql")
	controlLocation := filepath.Join(base, "steampipe_postgres_fdw.control")
	return sqlLocation, controlLocation
}

func getPostmasterPidLocation() string {
	return filepath.Join(getDataLocation(), "postmaster.pid")
}

func getPgHbaConfLocation() string {
	return filepath.Join(getDataLocation(), "pg_hba.conf")
}

func getPostgresqlConfLocation() string {
	return filepath.Join(getDataLocation(), "postgresql.conf")
}

func getPostgresqlConfDLocation() string {
	return filepath.Join(getDataLocation(), "postgresql.conf.d")
}

func getSteampipeConfLocation() string {
	return filepath.Join(getDataLocation(), "steampipe.conf")
}

func getLegacyPasswordFileLocation() string {
	return filepath.Join(getDatabaseLocation(), ".passwd")
}

func getPasswordFileLocation() string {
	return filepath.Join(filepaths.EnsureInternalDir(), ".passwd")
}

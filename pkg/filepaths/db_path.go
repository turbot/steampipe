package filepaths

import (
	"os"
	"path/filepath"

	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/platform"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
)

func ServiceExecutableRelativeLocation() string {
	return filepath.Join("db", constants.DatabaseVersion, "postgres", "bin", "postgres")
}

func DatabaseInstanceDir() string {
	loc := filepath.Join(EnsureDatabaseDir(), constants.DatabaseVersion)
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		err = os.MkdirAll(loc, 0755)
		error_helpers.FailOnErrorWithMessage(err, "could not create db version directory")
	}
	return loc
}

func GetDatabaseLocation() string {
	loc := filepath.Join(DatabaseInstanceDir(), "postgres")
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		err = os.MkdirAll(loc, 0755)
		error_helpers.FailOnErrorWithMessage(err, "could not create postgres installation directory")
	}
	return loc
}

func GetDataLocation() string {
	loc := filepath.Join(DatabaseInstanceDir(), "data")
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		err = os.MkdirAll(loc, 0755)
		error_helpers.FailOnErrorWithMessage(err, "could not create data directory")
	}
	return loc
}

// tar file where the dump file will be stored, so that it can be later restored after connections
// refresh in a new installation
func DatabaseBackupFilePath() string {
	return filepath.Join(EnsureDatabaseDir(), "backup.bk")
}

func GetDatabaseLibPath() string {
	return filepath.Join(GetDatabaseLocation(), "lib")
}

func GetRootCertLocation() string {
	return filepath.Join(GetDataLocation(), constants.RootCert)
}

func GetRootCertKeyLocation() string {
	return filepath.Join(GetDataLocation(), constants.RootCertKey)
}

func GetServerCertLocation() string {
	return filepath.Join(GetDataLocation(), constants.ServerCert)
}

func GetServerCertKeyLocation() string {
	return filepath.Join(GetDataLocation(), constants.ServerCertKey)
}

func GetInitDbBinaryExecutablePath() string {
	return filepath.Join(GetDatabaseLocation(), "bin", platform.Paths.InitDbExecutable)
}

func GetPostgresBinaryExecutablePath() string {
	return filepath.Join(GetDatabaseLocation(), "bin", platform.Paths.PostgresExecutable)
}

func PgDumpBinaryExecutablePath() string {
	return filepath.Join(GetDatabaseLocation(), "bin", platform.Paths.PgDumpExecutable)
}

func PgRestoreBinaryExecutablePath() string {
	return filepath.Join(GetDatabaseLocation(), "bin", platform.Paths.PgRestoreExecutable)
}

func GetDBSignatureLocation() string {
	loc := filepath.Join(GetDatabaseLocation(), "signature")
	return loc
}

func getDatabaseLibDirectory() string {
	return filepath.Join(GetDatabaseLocation(), "lib")
}

func GetFDWBinaryDir() string {
	return filepath.Join(getDatabaseLibDirectory(), "postgresql")
}

func GetFDWBinaryLocation() string {
	return filepath.Join(getDatabaseLibDirectory(), "postgresql", "steampipe_postgres_fdw.so")
}

func GetFDWSQLAndControlDir() string {
	return filepath.Join(GetDatabaseLocation(), "share", "postgresql", "extension")
}

func GetFDWSQLAndControlLocation() (string, string) {
	base := filepath.Join(GetDatabaseLocation(), "share", "postgresql", "extension")
	sqlLocation := filepath.Join(base, "steampipe_postgres_fdw--1.0.sql")
	controlLocation := filepath.Join(base, "steampipe_postgres_fdw.control")
	return sqlLocation, controlLocation
}

func GetPostmasterPidLocation() string {
	return filepath.Join(GetDataLocation(), "postmaster.pid")
}

func GetPgHbaConfLocation() string {
	return filepath.Join(GetDataLocation(), "pg_hba.conf")
}

func GetPostgresqlConfLocation() string {
	return filepath.Join(GetDataLocation(), "postgresql.conf")
}

func GetPostgresqlConfDLocation() string {
	return filepath.Join(GetDataLocation(), "postgresql.conf.d")
}

func GetSteampipeConfLocation() string {
	return filepath.Join(GetDataLocation(), "steampipe.conf")
}

func GetLegacyPasswordFileLocation() string {
	return filepath.Join(GetDatabaseLocation(), ".passwd")
}

func GetPasswordFileLocation() string {
	return filepath.Join(EnsureInternalDir(), ".passwd")
}

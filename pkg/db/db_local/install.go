package db_local

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/fatih/color"
	"github.com/jackc/pgx/v5"
	psutils "github.com/shirou/gopsutil/process"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	putils "github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
	"github.com/turbot/steampipe/v2/pkg/ociinstaller"
	"github.com/turbot/steampipe/v2/pkg/ociinstaller/versionfile"
	"github.com/turbot/steampipe/v2/pkg/statushooks"
)

var ensureMux sync.Mutex

// noBackupWarning is shown when we could not even take a backup (pg_dump) of
// the previous database before installing the new one. It is intentionally
// version-agnostic: the dump-side failure can happen on any upgrade and the
// previous version is not reliably known at this point.
func noBackupWarning() string {
	warningMessage := `Steampipe could not back up the data in your public schema before upgrading the embedded database.

Your previous data directory has been preserved under ~/.steampipe/db and was not modified. The new database has started with an empty public schema.

If you need that data, do not run another upgrade; open an issue at https://github.com/turbot/steampipe for recovery guidance.`

	return fmt.Sprintf("%s: %v\n", color.YellowString("Warning"), warningMessage)
}

// crossMajorPreflightSkippedWarning is shown on a cross-major upgrade (e.g.
// Postgres 14 -> 18) when the pre-flight collation scan detected
// collation-dependent objects in the public schema. Restoring those into a
// cluster whose default collation provider differs from the old one could
// silently corrupt index ordering and uniqueness, so the automatic restore is
// skipped. The new service starts fresh; the old data directory and a retained
// dump are kept so the user can migrate manually.
func crossMajorPreflightSkippedWarning(oldVersion, newVersion string) string {
	warningMessage := fmt.Sprintf(`The embedded database has been upgraded from PostgreSQL %s to PostgreSQL %s (a major version change).

Pre-flight detected collation-dependent objects (text indexes, unique constraints, or ordered views over non-ASCII data) in your public schema. A major-version upgrade changes the default collation provider, so automatically restoring these objects could silently corrupt their ordering or uniqueness. To keep the service usable and your data safe, the restore was skipped and the service started with an empty public schema.

Nothing was deleted. A dump of your old data has been retained in ~/.steampipe/backups and your previous database directory is preserved under ~/.steampipe/db. To restore manually, load the retained .sql dump into the new database, resolving any collation incompatibilities.`, oldVersion, newVersion)

	return fmt.Sprintf("%s: %v\n", color.YellowString("Warning"), warningMessage)
}

// crossMajorRestoreFailedWarning is shown on a cross-major upgrade when the
// best-effort automatic restore (pg_restore) returned a non-zero exit code.
// The new service starts fresh; the old data directory and a retained dump are
// kept so the user can migrate manually.
func crossMajorRestoreFailedWarning(oldVersion, newVersion string) string {
	warningMessage := fmt.Sprintf(`The embedded database has been upgraded from PostgreSQL %s to PostgreSQL %s (a major version change).

Steampipe attempted to automatically migrate your public schema, but pg_restore reported an error (some objects in the dump are not compatible with the new major version). To keep the service usable, it has started with an empty public schema.

Nothing was deleted. A dump of your old data has been retained in ~/.steampipe/backups and your previous database directory is preserved under ~/.steampipe/db. To restore manually, load the retained .sql dump into the new database, resolving any major-version incompatibilities.`, oldVersion, newVersion)

	return fmt.Sprintf("%s: %v\n", color.YellowString("Warning"), warningMessage)
}

// crossMajorValidationDivergedWarning is shown on a cross-major upgrade when
// pg_restore succeeded but the post-restore validation pass found the restored
// data diverged from the old cluster (row-count, sample-row checksum, or an
// invalid index). The restored schema is rolled back to an empty public schema
// so the service does not start with silently-corrupted data; the old data
// directory and a retained dump are kept so the user can migrate manually.
func crossMajorValidationDivergedWarning(oldVersion, newVersion string) string {
	warningMessage := fmt.Sprintf(`The embedded database has been upgraded from PostgreSQL %s to PostgreSQL %s (a major version change).

Steampipe automatically migrated your public schema, but a post-restore validation pass found the restored data diverged from your old database (a row count, a sample-row checksum, or an index validity check did not match). Rather than start with silently-corrupted data, the restore was rolled back and the service started with an empty public schema.

Nothing was deleted. A dump of your old data has been retained in ~/.steampipe/backups and your previous database directory is preserved under ~/.steampipe/db. To restore manually, load the retained .sql dump into the new database, resolving any collation incompatibilities.`, oldVersion, newVersion)

	return fmt.Sprintf("%s: %v\n", color.YellowString("Warning"), warningMessage)
}

// restoreFailedWarning is shown when a same-major (minor) migration took a
// backup but the automatic restore failed. The service is still allowed to
// start; the data is recoverable from the retained dump.
func restoreFailedWarning(newVersion string) string {
	warningMessage := fmt.Sprintf(`The embedded database was upgraded to PostgreSQL %s, but automatic restore of your public schema failed.

The service has started so it remains usable. Nothing was deleted: a dump of your old data has been retained in ~/.steampipe/backups and your previous database directory is preserved under ~/.steampipe/db. You can restore manually from the retained .sql dump, or open an issue at https://github.com/turbot/steampipe for help.`, newVersion)

	return fmt.Sprintf("%s: %v\n", color.YellowString("Warning"), warningMessage)
}

// dataTankMigrationDataPreservedWarning is the user/orchestrator-facing message
// emitted whenever the data-tank migration does not fully commit - a disk
// pre-flight or refresh-pause abort, a dump failure, a partial restore (tier 4
// reached but >=1 partition unmigrated), or all tiers exhausted. Under the
// 2026-06-08 governing decision (data-preservation over version-revert) the new
// Postgres version still runs, but the original is preserved on disk in two
// independent forms - the untouched old data directory plus the retained dump -
// so nothing is lost. The structured signal lives in the JSON marker file; this
// is the human-readable companion.
func dataTankMigrationDataPreservedWarning(retainedDumpPath string) string {
	warningMessage := fmt.Sprintf(`Data-tank migration to Postgres 18 could not be completed automatically.

The new database has started, but your original data has NOT been dropped - it is preserved on disk in two forms: your previous data directory under ~/.steampipe/db, and an insurance dump retained at %s.

This has been flagged for investigation. No action is required from you.`, retainedDumpPath)

	return fmt.Sprintf("%s: %v\n", color.YellowString("Warning"), warningMessage)
}

// Note: the per-cause data-tank fall-back warnings (refresh-pause-failed,
// disk-preflight-failed) were converged into dataTankMigrationDataPreservedWarning
// under the 2026-06-08 data-preservation decision. Every data-tank failure cause
// now surfaces the same outcome - the new version runs, the original is preserved
// on disk - so a single warning covers them all. The per-cause detail lives in the
// JSON marker file the engine writes (writeDataTankStatus).

// EnsureDBInstalled makes sure that the embedded postgres database is installed and ready to run
func EnsureDBInstalled(ctx context.Context) (err error) {
	putils.LogTime("db_local.EnsureDBInstalled start")

	ensureMux.Lock()

	doneChan := make(chan bool, 1)
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}

		putils.LogTime("db_local.EnsureDBInstalled end")
		ensureMux.Unlock()
		close(doneChan)
	}()

	if IsDBInstalled() {
		// check if the FDW need updating, and init the db if required
		err := prepareDb(ctx)
		return err
	}

	// handle the case that the previous db version may still be running
	dbState, err := GetState()
	if err != nil {
		log.Println("[TRACE] Error while loading database state", err)
		return err
	}
	if dbState != nil {
		return fmt.Errorf("cannot install service - a previous version of the Steampipe service is still running. To stop running services, use %s ", pconstants.Bold("steampipe service stop"))
	}

	log.Println("[TRACE] calling removeRunningInstanceInfo")
	err = removeRunningInstanceInfo()
	if err != nil && !os.IsNotExist(err) {
		log.Printf("[TRACE] removeRunningInstanceInfo failed: %v", err)
		return fmt.Errorf("Cleanup any Steampipe processes... FAILED!")
	}

	statushooks.SetStatus(ctx, "Installing database…")

	err = downloadAndInstallDbFiles(ctx)
	if err != nil {
		return err
	}

	statushooks.SetStatus(ctx, "Preparing backups…")

	// call prepareBackup to generate the db dump file if necessary
	// NOTE: this returns the existing database name - we use this when creating the new database
	dbName, err := prepareBackup(ctx)
	if err != nil {
		log.Printf("[ERROR] prepareBackup failed: %s", err.Error())
		if errors.Is(err, errDbInstanceRunning) {
			// remove the installation - otherwise, the backup won't get triggered, even if the user stops the service
			os.RemoveAll(filepaths.DatabaseInstanceDir())
			return err
		}
		// ignore all other errors with the backup, displaying a warning instead
		statushooks.Message(ctx, noBackupWarning())
	}

	// install the fdw
	_, err = installFDW(ctx, true)
	if err != nil {
		log.Printf("[TRACE] installFDW failed: %v", err)
		return fmt.Errorf("Download & install steampipe-postgres-fdw... FAILED!")
	}

	// run the database installation
	err = runInstall(ctx, dbName)
	if err != nil {
		return err
	}

	// write a signature after everything gets done!
	// so that we can check for this later on
	statushooks.SetStatus(ctx, "Updating install records…")
	err = updateDownloadedBinarySignature()
	if err != nil {
		log.Printf("[TRACE] updateDownloadedBinarySignature failed: %v", err)
		return fmt.Errorf("Updating install records... FAILED!")
	}

	return nil
}

func downloadAndInstallDbFiles(ctx context.Context) error {
	statushooks.SetStatus(ctx, "Prepare database install location…")
	// clear all db files
	err := os.RemoveAll(filepaths.GetDatabaseLocation())
	if err != nil {
		log.Printf("[TRACE] %v", err)
		return fmt.Errorf("Prepare database install location... FAILED!")
	}

	statushooks.SetStatus(ctx, "Download & install embedded PostgreSQL database…")
	_, err = ociinstaller.InstallDB(ctx, filepaths.GetDatabaseLocation())
	if err != nil {
		log.Printf("[TRACE] %v", err)
		return fmt.Errorf("Download & install embedded PostgreSQL database... FAILED!")
	}
	return nil
}

// IsDBInstalled checks and reports whether the embedded database binaries are available
func IsDBInstalled() bool {
	putils.LogTime("db_local.IsInstalled start")
	defer putils.LogTime("db_local.IsInstalled end")
	// check that both postgres binary and initdb binary exist
	if _, err := os.Stat(filepaths.GetInitDbBinaryExecutablePath()); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(filepaths.GetPostgresBinaryExecutablePath()); os.IsNotExist(err) {
		return false
	}
	return true
}

// IsFDWInstalled chceks whether all files required for the Steampipe FDW are available
func IsFDWInstalled() bool {
	fdwSQLFile, fdwControlFile := filepaths.GetFDWSQLAndControlLocation()
	if _, err := os.Stat(fdwSQLFile); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(fdwControlFile); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(filepaths.GetFDWBinaryLocation()); os.IsNotExist(err) {
		return false
	}
	return true
}

// prepareDb updates the db binaries and FDW if needed, and inits the database if required
func prepareDb(ctx context.Context) error {
	// load the db version info file
	putils.LogTime("db_local.LoadDatabaseVersionFile start")
	versionInfo, err := versionfile.LoadDatabaseVersionFile()
	putils.LogTime("db_local.LoadDatabaseVersionFile end")
	if err != nil {
		return err
	}

	// check if db needs to be updated
	// this means that the db version number has NOT changed but the package has changed
	// we can just drop in the new binaries
	if dbNeedsUpdate(versionInfo) {
		statushooks.SetStatus(ctx, "Updating database…")

		// install new db binaries
		if err = downloadAndInstallDbFiles(ctx); err != nil {
			return err
		}
		// write a signature after everything gets done!
		// so that we can check for this later on
		statushooks.SetStatus(ctx, "Updating install records…")
		if err = updateDownloadedBinarySignature(); err != nil {
			log.Printf("[TRACE] updateDownloadedBinarySignature failed: %v", err)
			return fmt.Errorf("Updating install records... FAILED!")
		}
	}

	// if the FDW is not installed, or needs an update
	if !IsFDWInstalled() || fdwNeedsUpdate(versionInfo) {
		// install fdw
		if _, err := installFDW(ctx, false); err != nil {
			log.Printf("[TRACE] installFDW failed: %v", err)
			return fmt.Errorf("Update steampipe-postgres-fdw... FAILED!")
		}

		// get the message renderer from the context
		// this allows the interactive client init to inject a custom renderer
		messageRenderer := statushooks.MessageRendererFromContext(ctx)
		messageRenderer("%s updated to %s.", pconstants.Bold("steampipe-postgres-fdw"), pconstants.Bold(constants.FdwVersion))
	}

	// Fire a migration when an old-version data directory is present even though
	// the binaries are already installed - the production (Pipes) case: the new PG
	// binaries are baked into the container image, so IsDBInstalled() is true and
	// the !IsDBInstalled() path (which normally runs prepareBackup) never executes,
	// yet a previous-version data dir is mounted in and must migrate. prepareBackup
	// dumps the old data to the backup file (a no-op when no old install is found);
	// restoreDBBackup loads it after the service starts. dbName is non-nil exactly
	// when a dump was taken, and on a cross-major jump prepareBackup leaves the OLD
	// cluster running for the restore/validation.
	dbName, backupErr := prepareBackup(ctx)
	if backupErr != nil {
		log.Printf("[ERROR] prepareBackup failed: %s", backupErr.Error())
		if errors.Is(backupErr, errDbInstanceRunning) {
			return backupErr
		}
		statushooks.Message(ctx, noBackupWarning())
		dbName = nil
	}
	migrationPending := dbName != nil

	if migrationPending {
		// runInstall wipes the current data dir contents (clearing any partial
		// cluster from a crashed previous attempt) and initdb's a fresh cluster;
		// the restore lands after the service starts. Do NOT kill instances here -
		// the old cluster prepareBackup started must stay live for restoreDBBackup.
		if err := runInstall(ctx, dbName); err != nil {
			return err
		}
	} else if needsInit() {
		statushooks.SetStatus(ctx, "Cleanup any Steampipe processes…")
		killInstanceIfAny(ctx)
		if err := runInstall(ctx, nil); err != nil {
			return err
		}
	}
	return nil
}

func fdwNeedsUpdate(versionInfo *versionfile.DatabaseVersionFile) bool {
	return versionInfo.FdwExtension.Version != constants.FdwVersion
}

func dbNeedsUpdate(versionInfo *versionfile.DatabaseVersionFile) bool {
	return versionInfo.EmbeddedDB.ImageDigest != constants.PostgresImageDigest
}

func installFDW(ctx context.Context, firstSetup bool) (string, error) {
	putils.LogTime("db_local.installFDW start")
	defer putils.LogTime("db_local.installFDW end")

	state, err := GetState()
	if err != nil {
		return "", err
	}
	if state != nil {
		defer func() {
			if !firstSetup {
				// update the signature
				updateDownloadedBinarySignature()
			}
		}()
	}
	statushooks.SetStatus(ctx, fmt.Sprintf("Download & install %s…", pconstants.Bold("steampipe-postgres-fdw")))
	return ociinstaller.InstallFdw(ctx, filepaths.GetDatabaseLocation())
}

func needsInit() bool {
	putils.LogTime("db_local.needsInit start")
	defer putils.LogTime("db_local.needsInit end")

	// test whether pg_hba.conf exists in our target directory
	return !filehelpers.FileExists(filepaths.GetPgHbaConfLocation())
}

func runInstall(ctx context.Context, oldDbName *string) error {
	putils.LogTime("db_local.runInstall start")
	defer putils.LogTime("db_local.runInstall end")

	statushooks.SetStatus(ctx, "Cleaning up…")

	err := putils.RemoveDirectoryContents(filepaths.GetDataLocation())
	if err != nil {
		log.Printf("[TRACE] %v", err)
		return fmt.Errorf("Prepare database install location... FAILED!")
	}

	statushooks.SetStatus(ctx, "Initializing database…")
	err = initDatabase()
	if err != nil {
		log.Printf("[TRACE] initDatabase failed: %v", err)
		return fmt.Errorf("Initializing database... FAILED!")
	}

	statushooks.SetStatus(ctx, "Starting database…")
	port, err := putils.GetNextFreePort()
	if err != nil {
		log.Printf("[TRACE] getNextFreePort failed: %v", err)
		return fmt.Errorf("Starting database... FAILED!")
	}

	process, err := startServiceForInstall(port)
	if err != nil {
		log.Printf("[TRACE] startServiceForInstall failed: %v", err)
		return fmt.Errorf("Starting database... FAILED!")
	}

	statushooks.SetStatus(ctx, "Connection to database…")
	client, err := createMaintenanceClient(ctx, port)
	if err != nil {
		return fmt.Errorf("Connection to database... FAILED!")
	}
	defer func() {
		statushooks.SetStatus(ctx, "Completing configuration")
		client.Close(ctx)
		doThreeStepPostgresExit(ctx, process)
	}()

	statushooks.SetStatus(ctx, "Generating database passwords…")
	// generate a password file for use later
	_, err = readPasswordFile()
	if err != nil {
		log.Printf("[TRACE] readPassword failed: %v", err)
		return fmt.Errorf("Generating database passwords... FAILED!")
	}

	// resolve the name of the database that is to be installed
	databaseName := resolveDatabaseName(oldDbName)
	// validate db name
	if !isValidDatabaseName(databaseName) {
		return fmt.Errorf("Invalid database name '%s' - must start with either a lowercase character or an underscore", databaseName)
	}

	statushooks.SetStatus(ctx, "Configuring database…")
	err = installDatabaseWithPermissions(ctx, databaseName, client)
	if err != nil {
		log.Printf("[TRACE] installSteampipeDatabaseAndUser failed: %v", err)
		return fmt.Errorf("Configuring database... FAILED!")
	}

	statushooks.SetStatus(ctx, "Configuring Steampipe…")
	err = installForeignServer(ctx, client)
	if err != nil {
		log.Printf("[TRACE] installForeignServer failed: %v", err)
		return fmt.Errorf("Configuring Steampipe... FAILED!")
	}

	return nil
}

func resolveDatabaseName(oldDbName *string) string {
	// resolve the name of the database that is to be installed
	// use the application constant as default
	if oldDbName != nil {
		return *oldDbName
	}
	databaseName := constants.DatabaseName
	if envValue, exists := os.LookupEnv(constants.EnvInstallDatabase); exists && len(envValue) > 0 {
		// use whatever is supplied, if available
		databaseName = envValue
	}
	return databaseName
}

func startServiceForInstall(port int) (*psutils.Process, error) {
	postgresCmd := exec.Command(
		filepaths.GetPostgresBinaryExecutablePath(),
		// by this time, we are sure that the port if free to listen to
		"-p", fmt.Sprint(port),
		"-c", "listen_addresses=127.0.0.1",
		// NOTE: If quoted, the application name includes the quotes. Worried about
		// having spaces in the APPNAME, but leaving it unquoted since currently
		// the APPNAME is hardcoded to be steampipe.
		"-c", fmt.Sprintf("application_name=%s", app_specific.AppName),
		"-c", fmt.Sprintf("cluster_name=%s", app_specific.AppName),

		// log directory
		"-c", fmt.Sprintf("log_directory=%s", filepaths.EnsureLogDir()),

		// Data Directory
		"-D", filepaths.GetDataLocation())

	setupLogCollection(postgresCmd)

	err := postgresCmd.Start()
	if err != nil {
		return nil, err
	}

	return psutils.NewProcess(int32(postgresCmd.Process.Pid))
}

func isValidDatabaseName(databaseName string) bool {
	if len(databaseName) == 0 {
		return false
	}
	return databaseName[0] == '_' || (databaseName[0] >= 'a' && databaseName[0] <= 'z')
}

func initDatabase() error {
	putils.LogTime("db_local.install.initDatabase start")
	defer putils.LogTime("db_local.install.initDatabase end")

	// initdb sometimes fail due to invalid locale settings, to avoid this we update
	// the locale settings to use 'C' only for the initdb process to complete, and
	// then return to the existing locale settings of the user.
	// set LC_ALL env variable to override current locale settings
	err := os.Setenv("LC_ALL", "C")
	if err != nil {
		log.Printf("[TRACE] failed to update locale settings:\n %s", err.Error())
		return err
	}

	initDBExecutable := filepaths.GetInitDbBinaryExecutablePath()
	initDbProcess := exec.Command(
		initDBExecutable,
		// Steampipe runs Postgres as a local, embedded database so trust local
		// users to login without a password.
		fmt.Sprintf("--auth=%s", "trust"),
		// Ensure the name of the database superuser is consistent across installs.
		// By default it would be based on the user running the install of this
		// embedded database.
		fmt.Sprintf("--username=%s", constants.DatabaseSuperUser),
		// Postgres data should placed under the Steampipe install directory.
		fmt.Sprintf("--pgdata=%s", filepaths.GetDataLocation()),
		// Ensure the encoding is consistent across installs. By default it would
		// be based on the system locale.
		fmt.Sprintf("--encoding=%s", "UTF-8"),
	)

	log.Printf("[TRACE] initdb start: %s", initDbProcess.String())

	output, runError := initDbProcess.CombinedOutput()
	if runError != nil {
		log.Printf("[TRACE] initdb failed:\n %s", string(output))
		return runError
	}

	// unset LC_ALL to return to original locale settings
	err = os.Unsetenv("LC_ALL")
	if err != nil {
		log.Printf("[TRACE] failed to return back to original locale settings:\n %s", err.Error())
		return err
	}

	// intentionally overwriting existing pg_hba.conf with a minimal config which only allows root
	// so that we can setup the database and permissions
	return os.WriteFile(filepaths.GetPgHbaConfLocation(), []byte(constants.MinimalPgHbaContent), 0600)
}

func installDatabaseWithPermissions(ctx context.Context, databaseName string, rawClient *pgx.Conn) error {
	putils.LogTime("db_local.install.installDatabaseWithPermissions start")
	defer putils.LogTime("db_local.install.installDatabaseWithPermissions end")

	log.Println("[TRACE] installing database with name", databaseName)

	statements := []string{

		// Lockdown all existing, and future, databases from use.
		`revoke all on database postgres from public`,
		`revoke all on database template1 from public`,

		// Only the root user (who owns the postgres database) should be able to use
		// or change it.
		`revoke all privileges on schema public from public`,

		// Create the steampipe database, used to hold all steampipe tables, views and data.
		fmt.Sprintf(`create database %s`, databaseName),

		// Restrict permissions from general users to the steampipe database. We add them
		// back progressively to allow appropriate read only access.
		fmt.Sprintf("revoke all on database %s from public", databaseName),

		// The root user gets full rights to the steampipe database, ensuring we can actually
		// configure and manage it properly.
		fmt.Sprintf("grant all on database %s to root", databaseName),

		// The root user gets a password which will be used later on to connect
		fmt.Sprintf(`alter user root with password '%s'`, generatePassword()),

		//
		// PERMISSIONS
		//
		// References:
		// * https://dba.stackexchange.com/questions/117109/how-to-manage-default-privileges-for-users-on-a-database-vs-schema/117661#117661
		//

		// Create a role to represent all steampipe_users in the database.
		// Grants and permissions can be managed on this role independent
		// of the actual users in the system, giving us flexibility.
		fmt.Sprintf(`create role %s`, constants.DatabaseUsersRole),

		// Allow the steampipe user access to the steampipe database only
		fmt.Sprintf("grant connect on database %s to %s", databaseName, constants.DatabaseUsersRole),

		// Create the steampipe user. By default they do not have superuser, createdb
		// or createrole permissions.
		fmt.Sprintf("create user %s", constants.DatabaseUser),

		// Allow the steampipe user to manage temporary tables
		fmt.Sprintf("grant temporary on database %s to %s", databaseName, constants.DatabaseUsersRole),

		// No need to set a password to the 'steampipe' user
		// The password gets set on every service start

		// Allow steampipe the privileges of steampipe_users.
		fmt.Sprintf("grant %s to %s", constants.DatabaseUsersRole, constants.DatabaseUser),
	}
	for _, statement := range statements {
		// not logging here, since the password may get logged
		// we don't want that
		if _, err := rawClient.Exec(ctx, statement); err != nil {
			return err
		}
	}
	return writePgHbaContent(databaseName, constants.DatabaseUser)
}

func writePgHbaContent(databaseName string, username string) error {
	content := fmt.Sprintf(constants.PgHbaTemplate, databaseName, username)
	return os.WriteFile(filepaths.GetPgHbaConfLocation(), []byte(content), 0600)
}

func installForeignServer(ctx context.Context, rawClient *pgx.Conn) error {
	putils.LogTime("db_local.installForeignServer start")
	defer putils.LogTime("db_local.installForeignServer end")

	statements := []string{
		// Install the FDW. The name must match the binary file.
		`drop extension if exists "steampipe_postgres_fdw" cascade`,
		`create extension if not exists "steampipe_postgres_fdw"`,
		// Use steampipe for the server name, it's simplest
		`create server "steampipe" foreign data wrapper "steampipe_postgres_fdw"`,
	}

	for _, statement := range statements {
		// NOTE: This may print a password to the log file, but it doesn't matter
		// since the password is stored in a config file anyway.
		log.Println("[TRACE] Install Foreign Server: ", statement)
		if _, err := rawClient.Exec(ctx, statement); err != nil {
			return err
		}
	}

	return nil
}

func updateDownloadedBinarySignature() error {
	putils.LogTime("db_local.updateDownloadedBinarySignature start")
	defer putils.LogTime("db_local.updateDownloadedBinarySignature end")

	versionInfo, err := versionfile.LoadDatabaseVersionFile()
	if err != nil {
		return err
	}
	installedSignature := fmt.Sprintf("%s|%s", versionInfo.EmbeddedDB.ImageDigest, versionInfo.FdwExtension.ImageDigest)
	return os.WriteFile(filepaths.GetDBSignatureLocation(), []byte(installedSignature), 0755)
}

package db_local

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/ociinstaller/versionfile"
	"github.com/turbot/steampipe/statushooks"
	"github.com/turbot/steampipe/utils"
)

var ensureMux sync.Mutex

// EnsureDBInstalled makes sure that the embedded pg database is installed and running
func EnsureDBInstalled(ctx context.Context) (err error) {
	utils.LogTime("db_local.EnsureDBInstalled start")

	ensureMux.Lock()

	doneChan := make(chan bool, 1)
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}

		utils.LogTime("db_local.EnsureDBInstalled end")
		ensureMux.Unlock()
		close(doneChan)
	}()

	if IsInstalled() {
		// check if the FDW need updating, and init the db id required
		err := prepareDb(ctx)
		return err
	}

	log.Println("[TRACE] calling removeRunningInstanceInfo")
	err = removeRunningInstanceInfo()
	if err != nil && !os.IsNotExist(err) {
		log.Printf("[TRACE] removeRunningInstanceInfo failed: %v", err)
		return fmt.Errorf("Cleanup any Steampipe processes... FAILED!")
	}

	log.Println("[TRACE] removing previous installation")
	statushooks.SetStatus(ctx, "Prepare database install location...")
	defer statushooks.Done(ctx)

	err = os.RemoveAll(getDatabaseLocation())
	if err != nil {
		log.Printf("[TRACE] %v", err)
		return fmt.Errorf("Prepare database install location... FAILED!")
	}

	statushooks.SetStatus(ctx, "Download & install embedded PostgreSQL database...")
	_, err = ociinstaller.InstallDB(ctx, constants.DefaultEmbeddedPostgresImage, getDatabaseLocation())
	if err != nil {
		log.Printf("[TRACE] %v", err)
		return fmt.Errorf("Download & install embedded PostgreSQL database... FAILED!")
	}

	_, err = installFDW(ctx, true)
	if err != nil {
		log.Printf("[TRACE] installFDW failed: %v", err)
		return fmt.Errorf("Download & install steampipe-postgres-fdw... FAILED!")
	}

	// run the database installation
	err = runInstall(ctx, true)
	if err != nil {
		return err
	}

	// write a signature after everything gets done!
	// so that we can check for this later on
	statushooks.SetStatus(ctx, "Updating install records...")
	err = updateDownloadedBinarySignature()
	if err != nil {
		log.Printf("[TRACE] updateDownloadedBinarySignature failed: %v", err)
		return fmt.Errorf("Updating install records... FAILED!")
	}

	return nil
}

// IsInstalled checks and reports whether the embedded database is installed and setup
func IsInstalled() bool {
	utils.LogTime("db_local.IsInstalled start")
	defer utils.LogTime("db_local.IsInstalled end")

	// check that both postgres binary and initdb binary exist
	// and are executable by us

	if _, err := os.Stat(getInitDbBinaryExecutablePath()); os.IsNotExist(err) {
		return false
	}

	if _, err := os.Stat(getPostgresBinaryExecutablePath()); os.IsNotExist(err) {
		return false
	}

	if _, err := os.Stat(getFDWBinaryLocation()); os.IsNotExist(err) {
		return false
	}

	fdwSQLFile, fdwControlFile := getFDWSQLAndControlLocation()

	if _, err := os.Stat(fdwSQLFile); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(fdwControlFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// prepareDb updates the FDW if needed, and inits the database if required
func prepareDb(ctx context.Context) error {
	// check if FDW needs to be updated
	if fdwNeedsUpdate() {
		_, err := installFDW(ctx, false)
		if err != nil {
			log.Printf("[TRACE] installFDW failed: %v", err)
			return fmt.Errorf("Update steampipe-postgres-fdw... FAILED!")
		}

		fmt.Printf("%s was updated to %s. ", constants.Bold("steampipe-postgres-fdw"), constants.Bold(constants.FdwVersion))
		fmt.Println()

	}

	if needsInit() {
		statushooks.SetStatus(ctx, "Cleanup any Steampipe processes...")
		killInstanceIfAny(ctx)
		if err := runInstall(ctx, false); err != nil {
			return err
		}
	}
	return nil
}

func fdwNeedsUpdate() bool {
	utils.LogTime("db_local.fdwNeedsUpdate start")
	defer utils.LogTime("db_local.fdwNeedsUpdate end")

	// check FDW version
	versionInfo, err := versionfile.LoadDatabaseVersionFile()
	if err != nil {
		utils.FailOnError(fmt.Errorf("could not verify required FDW version"))
	}
	return versionInfo.FdwExtension.Version != constants.FdwVersion
}

func installFDW(ctx context.Context, firstSetup bool) (string, error) {
	utils.LogTime("db_local.installFDW start")
	defer utils.LogTime("db_local.installFDW end")

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
	statushooks.SetStatus(ctx, fmt.Sprintf("Download & install %s...", constants.Bold("steampipe-postgres-fdw")))
	return ociinstaller.InstallFdw(ctx, constants.DefaultFdwImage, getDatabaseLocation())
}

func needsInit() bool {
	utils.LogTime("db_local.needsInit start")
	defer utils.LogTime("db_local.needsInit end")

	// test whether pg_hba.conf exists in our target directory
	return !helpers.FileExists(getPgHbaConfLocation())
}

func runInstall(ctx context.Context, firstInstall bool) error {
	utils.LogTime("db_local.runInstall start")
	defer utils.LogTime("db_local.runInstall end")

	statushooks.SetStatus(ctx, "Cleaning up...")
	defer statushooks.Done(ctx)

	err := utils.RemoveDirectoryContents(getDataLocation())
	if err != nil {
		log.Printf("[TRACE] %v", err)
		return fmt.Errorf("Prepare database install location... FAILED!")
	}

	statushooks.SetStatus(ctx, "Initializing database...")
	err = initDatabase()
	if err != nil {
		log.Printf("[TRACE] initDatabase failed: %v", err)
		return fmt.Errorf("Initializing database... FAILED!")
	}

	statushooks.SetStatus(ctx, "Starting database...")
	port, err := getNextFreePort()
	if err != nil {
		log.Printf("[TRACE] getNextFreePort failed: %v", err)
		return fmt.Errorf("Starting database... FAILED!")
	}

	process, err := startServiceForInstall(port)
	if err != nil {
		log.Printf("[TRACE] startServiceForInstall failed: %v", err)
		return fmt.Errorf("Starting database... FAILED!")
	}

	statushooks.SetStatus(ctx, "Connection to database...")
	client, err := createMaintenanceClient(ctx, port)
	if err != nil {
		return fmt.Errorf("Connection to database... FAILED!")
	}
	defer func() {
		statushooks.SetStatus(ctx, "Completing configuration")
		client.Close()
		doThreeStepPostgresExit(ctx, process)
	}()

	statushooks.SetStatus(ctx, "Generating database passwords...")
	// generate a password file for use later
	_, err = readPasswordFile()
	if err != nil {
		log.Printf("[TRACE] readPassword failed: %v", err)
		return fmt.Errorf("Generating database passwords... FAILED!")
	}

	// resolve the name of the database that is to be installed
	databaseName := resolveDatabaseName()
	// validate db name
	firstCharacter := databaseName[0:1]
	var ascii int
	for _, r := range databaseName {
		ascii = int(r)
		break
	}
	if firstCharacter == "_" || (ascii >= 'a' && ascii <= 'z') {
		log.Printf("[TRACE] valid database name: %s", databaseName)
	} else {
		return fmt.Errorf("Invalid database name '%s' - must start with either a lowercase character or an underscore", databaseName)
	}

	statushooks.SetStatus(ctx, "Configuring database...")
	err = installDatabaseWithPermissions(ctx, databaseName, client)
	if err != nil {
		log.Printf("[TRACE] installSteampipeDatabaseAndUser failed: %v", err)
		return fmt.Errorf("Configuring database... FAILED!")
	}

	statushooks.SetStatus(ctx, "Configuring Steampipe...")
	err = installForeignServer(ctx, client)
	if err != nil {
		log.Printf("[TRACE] installForeignServer failed: %v", err)
		return fmt.Errorf("Configuring Steampipe... FAILED!")
	}

	return nil
}

func resolveDatabaseName() string {
	// resolve the name of the database that is to be installed
	// use the application constant as default
	databaseName := constants.DatabaseName
	if envValue, exists := os.LookupEnv(constants.EnvInstallDatabase); exists && len(envValue) > 0 {
		// use whatever is supplied, if available
		databaseName = envValue
	}
	return databaseName
}

// createMaintenanceClient connects to the postgres server using the
// maintenance database and superuser
func createMaintenanceClient(ctx context.Context, port int) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=localhost port=%d user=%s dbname=postgres sslmode=disable", port, constants.DatabaseSuperUser)

	log.Println("[TRACE] Connection string: ", psqlInfo)

	// connect to the database using the postgres driver
	utils.LogTime("db_local.createClient connection open start")
	db, err := sql.Open("pgx", psqlInfo)
	db.SetMaxOpenConns(1)
	utils.LogTime("db_local.createClient connection open end")

	if err != nil {
		return nil, err
	}

	if err := db_common.WaitForConnection(ctx, db); err != nil {
		return nil, err
	}
	return db, nil
}

func startServiceForInstall(port int) (*psutils.Process, error) {
	postgresCmd := exec.Command(
		getPostgresBinaryExecutablePath(),
		// by this time, we are sure that the port if free to listen to
		"-p", fmt.Sprint(port),
		"-c", "listen_addresses=localhost",
		// NOTE: If quoted, the application name includes the quotes. Worried about
		// having spaces in the APPNAME, but leaving it unquoted since currently
		// the APPNAME is hardcoded to be steampipe.
		"-c", fmt.Sprintf("application_name=%s", constants.AppName),
		"-c", fmt.Sprintf("cluster_name=%s", constants.AppName),

		// log directory
		"-c", fmt.Sprintf("log_directory=%s", filepaths.LogDir()),

		// Data Directory
		"-D", getDataLocation())

	setupLogCollection(postgresCmd)

	err := postgresCmd.Start()
	if err != nil {
		return nil, err
	}

	return psutils.NewProcess(int32(postgresCmd.Process.Pid))
}

func getNextFreePort() (int, error) {
	utils.LogTime("db_local.install.getNextFreePort start")
	defer utils.LogTime("db_local.install.getNextFreePort end")
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return -1, err
	}
	defer listener.Close()
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return -1, fmt.Errorf("count not retrieve port")
	}
	return addr.Port, nil
}

func initDatabase() error {
	utils.LogTime("db_local.install.initDatabase start")
	defer utils.LogTime("db_local.install.initDatabase end")

	initDBExecutable := getInitDbBinaryExecutablePath()
	initDbProcess := exec.Command(
		initDBExecutable,
		fmt.Sprintf("--auth=%s", "trust"),
		fmt.Sprintf("--username=%s", constants.DatabaseSuperUser),
		fmt.Sprintf("--pgdata=%s", getDataLocation()),
		fmt.Sprintf("--encoding=%s", "UTF-8"),
		fmt.Sprintf("--wal-segsize=%d", 1),
		"--debug",
	)

	log.Printf("[TRACE] initdb start: %s", initDbProcess.String())

	output, runError := initDbProcess.CombinedOutput()
	if runError != nil {
		log.Printf("[TRACE] initdb failed:\n %s", string(output))
		return runError
	}

	// intentionally overwriting existing pg_hba.conf with a minimal config which only allows root
	// so that we can setup the database and permissions
	return os.WriteFile(getPgHbaConfLocation(), []byte(constants.MinimalPgHbaContent), 0600)
}

func installDatabaseWithPermissions(ctx context.Context, databaseName string, rawClient *sql.DB) error {
	utils.LogTime("db_local.install.installDatabaseWithPermissions start")
	defer utils.LogTime("db_local.install.installDatabaseWithPermissions end")

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
		if _, err := rawClient.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return writePgHbaContent(databaseName, constants.DatabaseUser)
}

func writePgHbaContent(databaseName string, username string) error {
	content := fmt.Sprintf(constants.PgHbaTemplate, databaseName, username)
	return os.WriteFile(getPgHbaConfLocation(), []byte(content), 0600)
}

func installForeignServer(ctx context.Context, rawClient *sql.DB) error {
	utils.LogTime("db_local.installForeignServer start")
	defer utils.LogTime("db_local.installForeignServer end")

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
		if _, err := rawClient.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	return nil
}

func updateDownloadedBinarySignature() error {
	utils.LogTime("db_local.updateDownloadedBinarySignature start")
	defer utils.LogTime("db_local.updateDownloadedBinarySignature end")

	versionInfo, err := versionfile.LoadDatabaseVersionFile()
	if err != nil {
		return err
	}
	installedSignature := fmt.Sprintf("%s|%s", versionInfo.EmbeddedDB.ImageDigest, versionInfo.FdwExtension.ImageDigest)
	return os.WriteFile(getDBSignatureLocation(), []byte(installedSignature), 0755)
}

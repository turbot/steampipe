package db_local

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/briandowns/spinner"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/ociinstaller/versionfile"
	"github.com/turbot/steampipe/utils"
)

var ensureMux sync.Mutex

// EnsureDBInstalled makes sure that the embedded pg database is installed and running
func EnsureDBInstalled() (err error) {
	utils.LogTime("db.EnsureDBInstalled start")

	ensureMux.Lock()

	doneChan := make(chan bool, 1)
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}

		utils.LogTime("db.EnsureDBInstalled end")
		ensureMux.Unlock()
		close(doneChan)
	}()

	spinner := display.StartSpinnerAfterDelay("", constants.SpinnerShowTimeout, doneChan)

	if IsInstalled() {
		// check if the FDW need updating, and init the db id required
		err := PrepareDb(spinner)
		display.StopSpinner(spinner)
		return err
	}

	log.Println("[TRACE] calling killPreviousInstanceIfAny")
	display.UpdateSpinnerMessage(spinner, "Cleanup any Steampipe processes...")
	killInstanceIfAny()
	log.Println("[TRACE] calling removeRunningInstanceInfo")
	err = removeRunningInstanceInfo()
	if err != nil && !os.IsNotExist(err) {
		display.StopSpinner(spinner)
		log.Printf("[TRACE] removeRunningInstanceInfo failed: %v", err)
		return fmt.Errorf("Cleanup any Steampipe processes... FAILED!")
	}

	log.Println("[TRACE] removing previous installation")
	display.UpdateSpinnerMessage(spinner, "Prepare database install location...")
	err = os.RemoveAll(getDatabaseLocation())
	if err != nil {
		display.StopSpinner(spinner)
		log.Printf("[TRACE] %v", err)
		return fmt.Errorf("Prepare database install location... FAILED!")
	}

	display.UpdateSpinnerMessage(spinner, "Download & install embedded PostgreSQL database...")
	_, err = ociinstaller.InstallDB(constants.DefaultEmbeddedPostgresImage, getDatabaseLocation())
	if err != nil {
		display.StopSpinner(spinner)
		log.Printf("[TRACE] %v", err)
		return fmt.Errorf("Download & install embedded PostgreSQL database... FAILED!")
	}

	// installFDW takes care of the spinner, since it may need to run independently
	_, err = installFDW(true, spinner)
	if err != nil {
		display.StopSpinner(spinner)
		log.Printf("[TRACE] installFDW failed: %v", err)
		return fmt.Errorf("Download & install steampipe-postgres-fdw... FAILED!")
	}

	// do init
	err = doInit(true, spinner)
	if err != nil {
		display.StopSpinner(spinner)
		return err
	}

	// write a signature after everything gets done!
	// so that we can check for this later on
	display.UpdateSpinnerMessage(spinner, "Updating install records...")
	err = updateDownloadedBinarySignature()
	if err != nil {
		display.StopSpinner(spinner)
		log.Printf("[TRACE] updateDownloadedBinarySignature failed: %v", err)
		return fmt.Errorf("Updating install records... FAILED!")
	}

	display.StopSpinner(spinner)
	return nil
}

// PrepareDb updates the FDW if needed, and inits the database if required
func PrepareDb(spinner *spinner.Spinner) error {
	// check if FDW needs to be updated
	if fdwNeedsUpdate() {
		_, err := installFDW(false, spinner)
		spinner.Stop()
		if err != nil {
			log.Printf("[TRACE] installFDW failed: %v", err)
			return fmt.Errorf("Update steampipe-postgres-fdw... FAILED!")
		}

		fmt.Printf("%s was updated to %s. ", constants.Bold("steampipe-postgres-fdw"), constants.Bold(constants.FdwVersion))
		fmt.Println()

	}

	if needsInit() {
		spinner.Start()
		display.UpdateSpinnerMessage(spinner, "Cleanup any Steampipe processes...")
		killInstanceIfAny()
		if err := doInit(false, spinner); err != nil {
			return err
		}
	}
	return nil
}

// IsInstalled checks and reports whether the embedded database is installed and setup
func IsInstalled() bool {
	utils.LogTime("db.IsInstalled start")
	defer utils.LogTime("db.IsInstalled end")

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

	// TO DO:
	// 	- move this out ot this function - make the "upgrade" optional
	// 	- add function to get the latest digest
	// // get the signature of the binary
	// // we do this by having a signature file
	// // which stores the md5 hash of the URL we downloaded
	// // from and then comparing that against the has of the
	// // URL we have  in the constants
	// installedSignature := getInstalledBinarySignature()
	// desiredSignature := get

	// if installedSignature != desiredSignature {
	// 	return true
	// }

	return true
}

func fdwNeedsUpdate() bool {
	utils.LogTime("db.fdwNeedsUpdate start")
	defer utils.LogTime("db.fdwNeedsUpdate end")

	// check FDW version
	versionInfo, err := versionfile.LoadDatabaseVersionFile()
	if err != nil {
		utils.FailOnError(fmt.Errorf("could not verify required FDW version"))
	}
	return versionInfo.FdwExtension.Version != constants.FdwVersion
}

func installFDW(firstSetup bool, spinner *spinner.Spinner) (string, error) {
	utils.LogTime("db.installFDW start")
	defer utils.LogTime("db.installFDW end")

	status, err := GetStatus()
	if err != nil {
		return "", err
	}
	if status != nil {
		defer func() {
			if !firstSetup {
				// update the signature
				updateDownloadedBinarySignature()
			}
		}()
	}
	display.UpdateSpinnerMessage(spinner, fmt.Sprintf("Download & install %s...", constants.Bold("steampipe-postgres-fdw")))
	return ociinstaller.InstallFdw(constants.DefaultFdwImage, getDatabaseLocation())
}

func needsInit() bool {
	utils.LogTime("db.needsInit start")
	defer utils.LogTime("db.needsInit end")

	// test whether pg_hba.conf exists in our target directory
	return !helpers.FileExists(getPgHbaConfLocation())
}

func doInit(firstInstall bool, spinner *spinner.Spinner) error {
	utils.LogTime("db.doInit start")
	defer utils.LogTime("db.doInit end")

	display.UpdateSpinnerMessage(spinner, "Cleaning up...")
	err := utils.RemoveDirectoryContents(getDataLocation())
	if err != nil {
		display.StopSpinner(spinner)
		log.Printf("[TRACE] %v", err)
		return fmt.Errorf("Prepare database install location... FAILED!")
	}

	display.UpdateSpinnerMessage(spinner, "Initializing database...")
	err = initDatabase()
	if err != nil {
		display.StopSpinner(spinner)
		log.Printf("[TRACE] initDatabase failed: %v", err)
		return fmt.Errorf("Initializing database... FAILED!")
	}

	display.UpdateSpinnerMessage(spinner, "Generating database passwords...")
	steampipePassword, err := readPasswordFile()
	if err != nil {
		display.StopSpinner(spinner)
		log.Printf("[TRACE] readPassword failed: %v", err)
		return fmt.Errorf("Generating database passwords... FAILED!")
	}
	// we will generate and lock with an ephemereal root password
	rootPassword := generatePassword()

	display.UpdateSpinnerMessage(spinner, "Starting database...")
	err = startPostgresProcessAndSetPassword(constants.DatabaseDefaultPort, ListenTypeLocal, constants.InvokerInstaller)
	if err != nil {
		display.StopSpinner(spinner)
		log.Printf("[TRACE] startPostgresProcess failed: %v", err)
		return fmt.Errorf("Starting database... FAILED!")
	}

	display.UpdateSpinnerMessage(spinner, "Configuring database...")
	err = installSteampipeDatabaseAndUser(steampipePassword, rootPassword)
	if err != nil {
		display.StopSpinner(spinner)
		log.Printf("[TRACE] installSteampipeDatabaseAndUser failed: %v", err)
		return fmt.Errorf("Configuring database... FAILED!")
	}

	display.UpdateSpinnerMessage(spinner, "Configuring Steampipe...")
	err = installForeignServer()
	if err != nil {
		display.StopSpinner(spinner)
		log.Printf("[TRACE] installForeignServer failed: %v", err)
		return fmt.Errorf("Configuring Steampipe... FAILED!")
	}
	// force stop
	display.UpdateSpinnerMessage(spinner, "Completing configuration")
	_, err = StopDB(false, constants.InvokerInstaller, nil)

	return err
}

func initDatabase() error {
	utils.LogTime("db.initDatabase start")
	defer utils.LogTime("db.initDatabase end")

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

	// intentionally overwriting existing pg_hba.conf
	return ioutil.WriteFile(getPgHbaConfLocation(), []byte(constants.PgHbaContent), 0600)
}

func installSteampipeDatabaseAndUser(steampipePassword string, rootPassword string) error {
	utils.LogTime("db.installSteampipeDatabase start")
	defer utils.LogTime("db.installSteampipeDatabase end")

	rawClient, err := createLocalDbClient("postgres", constants.DatabaseSuperUser)
	if err != nil {
		return err
	}

	defer func() {
		rawClient.Close()
	}()

	statements := []string{

		// Lockdown all existing, and future, databases from use.
		`revoke all on database postgres from public`,
		`revoke all on database template1 from public`,

		// Only the root user (who owns the postgres database) should be able to use
		// or change it.
		`revoke all privileges on schema public from public`,

		// Create the steampipe database, used to hold all steampipe tables, views and data.
		`create database steampipe`,

		// Restrict permissions from general users to the steampipe database. We add them
		// back progressively to allow appropriate read only access.
		`revoke all on database steampipe from public`,

		// The root user gets full rights to the steampipe database, ensuring we can actually
		// configure and manage it properly.
		`grant all on database steampipe to root`,

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
		`create role steampipe_users`,

		// Allow the steampipe user access to the steampipe database only
		`grant connect on database steampipe to steampipe_users`,

		// Create the steampipe user. By default they do not have superuser, createdb
		// or createrole permissions.
		`create user steampipe`,

		// Allow the steampipe user to manage temporary tables
		`grant temporary on database steampipe to steampipe_users`,

		// No need to set a password to the 'steampipe' user
		// The password gets set on every service start

		// Allow steampipe the privileges of steampipe_users.
		`grant steampipe_users to steampipe`,
	}

	for _, statement := range statements {
		// NOTE: This may print a password to the log file, but it doesn't matter
		// since the password is stored in a config file anyway.
		log.Println("[TRACE] Install steampipe database: ", statement)
		if _, err := rawClient.Exec(statement); err != nil {
			return err
		}
	}

	return nil
}

func installForeignServer() error {
	utils.LogTime("db.installForeignServer start")
	defer utils.LogTime("db.installForeignServer end")

	statements := []string{
		// Install the FDW. The name must match the binary file.
		`drop extension if exists "steampipe_postgres_fdw" cascade`,
		`create extension if not exists "steampipe_postgres_fdw"`,
		// Use steampipe for the server name, it's simplest
		`create server "steampipe" foreign data wrapper "steampipe_postgres_fdw"`,
	}
	_, err := executeSqlAsRoot(statements...)
	return err
}

func updateDownloadedBinarySignature() error {
	utils.LogTime("db.updateDownloadedBinarySignature start")
	defer utils.LogTime("db.updateDownloadedBinarySignature end")

	versionInfo, err := versionfile.LoadDatabaseVersionFile()
	if err != nil {
		return err
	}
	installedSignature := fmt.Sprintf("%s|%s", versionInfo.EmbeddedDB.ImageDigest, versionInfo.FdwExtension.ImageDigest)
	return ioutil.WriteFile(getDBSignatureLocation(), []byte(installedSignature), 0755)
}

func getInstalledBinarySignature() string {
	sigBytes, err := ioutil.ReadFile(getDBSignatureLocation())
	sig := ""
	if os.IsNotExist(err) {
		sig = ""
	} else {
		sig = string(sigBytes)
	}
	return sig
}

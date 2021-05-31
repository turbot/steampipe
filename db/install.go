package db

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/ociinstaller/versionfile"
	"github.com/turbot/steampipe/utils"
)

var ensureMux sync.Mutex

// EnsureDBInstalled makes sure that the embedded pg database is installed and running
func EnsureDBInstalled() {
	utils.LogTime("setup start")
	log.Println("[TRACE] ensure installed")

	ensureMux.Lock()
	defer func() {
		log.Println("[TRACE] ensured installed")
		utils.LogTime("setup end")
		ensureMux.Unlock()
	}()

	if IsInstalled() {
		// check if FDW needs to be updated
		if fdwNeedsUpdate() {
			_, err := installFDW(false)
			if err != nil {
				utils.ShowError(fmt.Errorf("%s could not be updated", constants.Bold("steampipe-postgres-fdw")))
			} else {
				fmt.Printf("%s was updated to %s. ", constants.Bold("steampipe-postgres-fdw"), constants.Bold(constants.FdwVersion))
				currentStatus, err := GetStatus()
				if err != nil || currentStatus != nil {
					fmt.Printf("Run %s for change to take effect.", constants.Bold("steampipe service restart"))
				}
				fmt.Println()
			}
		}
		return
	}

	log.Println("[TRACE] calling killPreviousInstanceIfAny")
	killPreviousSpinner := display.ShowSpinner(fmt.Sprintf("Cleanup any Steampipe processes..."))
	killPreviousInstanceIfAny()
	log.Println("[TRACE] calling removeRunningInstanceInfo")
	err := removeRunningInstanceInfo()
	display.StopSpinner(killPreviousSpinner)
	if err != nil && !os.IsNotExist(err) {
		utils.FailOnErrorWithMessage(err, "x Cleanup any Steampipe processes... FAILED!")
	}

	log.Println("[TRACE] removing previous installation")
	dbCleanupSpinner := display.ShowSpinner(fmt.Sprintf("Prepare database install location..."))
	err = os.RemoveAll(getDatabaseLocation())
	if err != nil {
		display.StopSpinner(dbCleanupSpinner)
		utils.FailOnErrorWithMessage(err, "x Prepare database install location... FAILED!")
	}
	err = os.RemoveAll(getDataLocation())
	if err != nil {
		display.StopSpinner(dbCleanupSpinner)
		utils.FailOnErrorWithMessage(err, "x Prepare database install location... FAILED!")
	}
	display.StopSpinner(dbCleanupSpinner)

	dbInstallSpinner := display.ShowSpinner(fmt.Sprintf("Download & install embedded PostgreSQL database..."))
	_, err = ociinstaller.InstallDB(constants.DefaultEmbeddedPostgresImage, getDatabaseLocation())
	display.StopSpinner(dbInstallSpinner)
	if err != nil {
		utils.FailOnErrorWithMessage(err, "x Download & install embedded PostgreSQL database... FAILED!")
	}

	// installFDW takes care of the spinner, since it may need to run independently
	_, err = installFDW(true)

	dbInitSpinner := display.ShowSpinner(fmt.Sprintf("Initializing database..."))
	err = initDatabase()
	display.StopSpinner(dbInitSpinner)
	if err != nil {
		utils.FailOnErrorWithMessage(err, "x Initializing database... FAILED!")
	}

	pwSpinner := display.ShowSpinner(fmt.Sprintf("Generating database passwords..."))
	// Try for passwords of the form dbC-3Ji-d04d
	steampipePassword := generatePassword()
	rootPassword := generatePassword()
	// write the passwords that were generated
	err = writePasswordFile(steampipePassword, rootPassword)
	display.StopSpinner(pwSpinner)
	if err != nil {
		utils.FailOnErrorWithMessage(err, "x Generating database passwords... FAILED!")
	}

	startServiceSpinner := display.ShowSpinner(fmt.Sprintf("Configuring database..."))
	StartService(InvokerInstaller)
	err = installSteampipeDatabase(steampipePassword, rootPassword)
	display.StopSpinner(startServiceSpinner)
	if err != nil {
		utils.FailOnErrorWithMessage(err, "x Configuring database... FAILED!")
	}

	installSteampipeSpinner := display.ShowSpinner(fmt.Sprintf("Configuring Steampipe..."))
	err = installSteampipeHub()
	display.StopSpinner(installSteampipeSpinner)
	if err != nil {
		utils.FailOnErrorWithMessage(err, "x Configuring Steampipe... FAILED!")
	}
	// force stop
	StopDB(true, InvokerInstaller)

	// write a signature after everything gets done!
	// so that we can check for this later on
	writeSignaturesSpinner := display.ShowSpinner(fmt.Sprintf("Updating install records..."))
	err = updateDownloadedBinarySignature()
	display.StopSpinner(writeSignaturesSpinner)
	if err != nil {
		utils.FailOnErrorWithMessage(err, "x Updating install records... FAILED!")
	}
}

func installSteampipeDatabase(withSteampipePassword string, withRootPassword string) error {
	rawClient, err := createPostgresDbClient()
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
		fmt.Sprintf(`alter user root with password '%s'`, withRootPassword),

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

		// Set a random, complex password for the steampipe user. Done as a separate
		// step from the create for clarity and reuse.
		// TODO: need a complex random password here, that is available for sharing with the user when the do steampipe service
		fmt.Sprintf(`alter user steampipe with password '%s'`, withSteampipePassword),

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

func installSteampipeHub() error {
	statements := []string{
		// Install the FDW. The name must match the binary file.
		`drop extension if exists "steampipe_postgres_fdw" cascade`,
		`create extension if not exists "steampipe_postgres_fdw"`,
		// Use steampipe for the server name, it's simplest
		`create server "steampipe" foreign data wrapper "steampipe_postgres_fdw"`,
	}
	_, err := executeSqlAsRoot(statements)
	return err
}

func fdwNeedsUpdate() bool {
	// check FDW version
	versionInfo, err := versionfile.Load()
	if err != nil {
		utils.FailOnError(fmt.Errorf("could not verify required FDW version"))
	}
	return versionInfo.FdwExtension.Version != constants.FdwVersion
}

func installFDW(firstSetup bool) (string, error) {
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
	fdwInstallSpinner := display.ShowSpinner(fmt.Sprintf("Download & install %s...", constants.Bold("steampipe-postgres-fdw")))
	newDigest, err := ociinstaller.InstallFdw(constants.DefaultFdwImage, getDatabaseLocation())
	display.StopSpinner(fdwInstallSpinner)
	if err != nil {
		if firstSetup {
			utils.FailOnErrorWithMessage(err, "x Download & install steampipe-postgres-fdw... FAILED!")
		} else {
			utils.ShowErrorWithMessage(err, "could not update steampipe-postgres-fdw")
		}
	}
	return newDigest, err
}

// IsInstalled :: checks and reports whether the embedded database is installed and setup
func IsInstalled() bool {
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

func updateDownloadedBinarySignature() error {
	versionInfo, err := versionfile.Load()
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

func initDatabase() error {
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

	log.Printf("[TRACE] %s", initDbProcess.String())

	runError := initDbProcess.Run()
	if runError != nil {
		return runError
	}

	return ioutil.WriteFile(getPgHbaConfLocation(), []byte(constants.PgHbaContent), 0600)
}

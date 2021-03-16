package db

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/ociinstaller/versionfile"
	"github.com/turbot/steampipe/utils"
)

var ensureMux sync.Mutex

// EnsureDBInstalled makes sure that the embedded pg database is installed and running
func EnsureDBInstalled() {
	logging.LogTime("setup start")
	log.Println("[TRACE] ensure installed")

	ensureMux.Lock()
	defer func() {
		log.Println("[TRACE] ensured installed")
		logging.LogTime("setup end")
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
	killPreviousSpinner := utils.ShowSpinner(fmt.Sprintf("Cleanup any Steampipe processes..."))
	killPreviousInstanceIfAny()
	log.Println("[TRACE] calling removeRunningInstanceInfo")
	err := removeRunningInstanceInfo()
	utils.StopSpinner(killPreviousSpinner)
	if err != nil && !os.IsNotExist(err) {
		utils.FailOnErrorWithMessage(err, "x Cleanup any Steampipe processes... FAILED!")
	}

	log.Println("[TRACE] removing previous installation")
	dbCleanupSpinner := utils.ShowSpinner(fmt.Sprintf("Prepare database install location..."))
	err = os.RemoveAll(getDatabaseLocation())
	if err != nil {
		utils.StopSpinner(dbCleanupSpinner)
		utils.FailOnErrorWithMessage(err, "x Prepare database install location... FAILED!")
	}
	err = os.RemoveAll(getDataLocation())
	if err != nil {
		utils.StopSpinner(dbCleanupSpinner)
		utils.FailOnErrorWithMessage(err, "x Prepare database install location... FAILED!")
	}
	utils.StopSpinner(dbCleanupSpinner)

	dbInstallSpinner := utils.ShowSpinner(fmt.Sprintf("Download & install embedded PostgreSQL database..."))
	_, err = ociinstaller.InstallDB(constants.DefaultEmbeddedPostgresImage, getDatabaseLocation())
	utils.StopSpinner(dbInstallSpinner)
	if err != nil {
		utils.FailOnErrorWithMessage(err, "x Download & install embedded PostgreSQL database... FAILED!")
	}

	// installFDW takes care of the spinner, since it may need to run independently
	_, err = installFDW(true)

	dbInitSpinner := utils.ShowSpinner(fmt.Sprintf("Initializing database..."))
	err = initDatabase()
	utils.StopSpinner(dbInitSpinner)
	if err != nil {
		utils.FailOnErrorWithMessage(err, "x Initializing database... FAILED!")
	}

	pwSpinner := utils.ShowSpinner(fmt.Sprintf("Generating database passwords..."))
	// Try for passwords of the form dbC-3Ji-d04d
	steampipePassword := generatePassword()
	rootPassword := generatePassword()
	// write the passwords that were generated
	err = writePasswordFile(steampipePassword, rootPassword)
	utils.StopSpinner(pwSpinner)
	if err != nil {
		utils.FailOnErrorWithMessage(err, "x Generating database passwords... FAILED!")
	}

	startServiceSpinner := utils.ShowSpinner(fmt.Sprintf("Configuring database..."))
	StartService(InvokerInstaller)
	defer func() {
		// force stop
		StopDB(true)
	}()
	err = installSteampipeDatabase(steampipePassword, rootPassword)
	utils.StopSpinner(startServiceSpinner)
	if err != nil {
		utils.FailOnErrorWithMessage(err, "x Configuring database... FAILED!")
	}

	installSteampipeSpinner := utils.ShowSpinner(fmt.Sprintf("Configuring Steampipe..."))
	err = installSteampipeHub()
	utils.StopSpinner(installSteampipeSpinner)
	if err != nil {
		utils.FailOnErrorWithMessage(err, "x Configuring Steampipe... FAILED!")
	}

	// write a signature after everything gets done!
	// so that we can check for this later on
	writeSignaturesSpinner := utils.ShowSpinner(fmt.Sprintf("Updating install records..."))
	err = updateDownloadedBinarySignature()
	utils.StopSpinner(writeSignaturesSpinner)
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
	rawClient, err := createSteampipeRootDbClient()
	if err != nil {
		return err
	}

	defer func() {
		rawClient.Close()
	}()

	statements := []string{
		// Install the FDW. The name must match the binary file.
		`drop extension if exists "steampipe_postgres_fdw" cascade`,
		`create extension if not exists "steampipe_postgres_fdw"`,
		// Use steampipe for the server name, it's simplest
		`create server "steampipe" foreign data wrapper "steampipe_postgres_fdw"`,
	}

	for _, statement := range statements {
		log.Println("[TRACE] Install steampipe hub: ", statement)
		if _, err := rawClient.Exec(statement); err != nil {
			return err
		}
	}

	return nil
}

// StartService :: invokes `steampipe service start --listen local --refresh=false --invoker query`
func StartService(invoker Invoker) {
	log.Println("[TRACE] start service")
	// spawn a process to start the service, passing refresh=false to ensure we DO NOT refresh connections
	// (as we will do that ourselves)
	cmd := exec.Command(os.Args[0], "service", "start", "--listen", "local", "--refresh=false", "--invoker", string(invoker), "--install-dir", constants.SteampipeDir)
	cmd.Start()
	startedAt := time.Now()
	spinnerShown := false
	startedChannel := make(chan bool, 1)

	go func() {
		for {
			st, err := loadRunningInstanceInfo()
			if err != nil {
				utils.ShowError(errors.New("could not retrieve service status"))
				return
			}
			if st != nil {
				startedChannel <- true
				break
			}
			if time.Since(startedAt) > constants.SpinnerShowTimeout && !spinnerShown {
				if cmdconfig.Viper().GetBool(constants.ShowInteractiveOutputConfigKey) {
					s := utils.ShowSpinner("Waiting for database to start...")
					defer utils.StopSpinner(s)
				}
				// set this anyway, so that next time it doesn't come in
				spinnerShown = true
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()

	timeoutAfter := time.After(1 * time.Minute)

	select {
	case <-timeoutAfter:
		utils.ShowError(fmt.Errorf("x Waiting for database to start... FAILED!"))
		return
	case <-startedChannel:
		// do nothing
	}
	return
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
	fdwInstallSpinner := utils.ShowSpinner(fmt.Sprintf("Download & install %s...", constants.Bold("steampipe-postgres-fdw")))
	newDigest, err := ociinstaller.InstallFdw(constants.DefaultFdwImage, getDatabaseLocation())
	utils.StopSpinner(fdwInstallSpinner)
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

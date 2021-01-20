package db

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"sync"

	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/ociinstaller"
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
		return
	}

	log.Println("[TRACE] calling killPreviousInstanceIfAny")
	killPreviousInstanceIfAny()
	log.Println("[TRACE] calling removeRunningInstanceInfo")
	err := removeRunningInstanceInfo()
	if err != nil && !os.IsNotExist(err) {
		utils.FailOnErrorWithMessage(err, "Installation failed")
	}

	log.Println("[TRACE] removing previous installation")
	err = os.RemoveAll(getDatabaseLocation())
	if err != nil {
		utils.FailOnErrorWithMessage(err, "Installation failed")
	}
	err = os.RemoveAll(getDataLocation())
	if err != nil {
		utils.FailOnErrorWithMessage(err, "Installation failed")
	}

	fmt.Printf("\nInstalling dbClient from image: %s\n", constants.DefaultEmbeddedPostgresImage)
	dbImageDigest, err := ociinstaller.InstallDB(constants.DefaultEmbeddedPostgresImage, getDatabaseLocation())
	utils.FailOnErrorWithMessage(err, "dbClient Installation failed")

	fmt.Printf("\nInstalling hub extension from image: %s\n", constants.DefaultFdwImage)
	fdwDigest, err := ociinstaller.InstallFdw(constants.DefaultFdwImage, getDatabaseLocation())
	utils.FailOnErrorWithMessage(err, "Hub extension installation failed")

	fmt.Println("Initializing SQL Support...")
	err = initDatabase()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	utils.FailOnError(err)

	StartService(InvokerInstaller)

	defer func() {
		// force stop
		StopDB(true)
	}()

	fmt.Println("Initializing steampipe Database...")
	err = installSteampipeDatabase()
	utils.FailOnErrorWithMessage(err, "failed")

	fmt.Println("Installing SteampipeHub...")
	err = installSteampipeHub()
	utils.FailOnErrorWithMessage(err, "failed")

	// write a signature after everything gets done!
	// so that we can check for this later on
	writeDownloadedBinarySignature(dbImageDigest, fdwDigest)
}

func installSteampipeDatabase() error {
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
		`alter user steampipe with password 'password001'`,

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
	cmd := exec.Command(os.Args[0], "service", "start", "--listen", "local", "--refresh=false", "--invoker", string(invoker))
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
				s := utils.ShowSpinner("Waiting for service startup")
				spinnerShown = true
				defer utils.StopSpinner(s)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()

	timeoutAfter := time.After(1 * time.Minute)

	select {
	case <-timeoutAfter:
		utils.ShowError(fmt.Errorf("could not start service"))
		return
	case <-startedChannel:
		// do nothing
	}
	return
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

func writeDownloadedBinarySignature(dbDigest string, fdwDigest string) {
	installedSignature := fmt.Sprintf("%s|%s", dbDigest, fdwDigest)
	ioutil.WriteFile(getDBSignatureLocation(), []byte(installedSignature), 0755)
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

	return initDbProcess.Run()
}

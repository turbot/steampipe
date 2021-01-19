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

	fmt.Println("Initialising SQL Support...")
	err = initDatabase()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	utils.FailOnError(err)

	fmt.Println("Installing SteampipeHub...")
	err = installSteampipeHub()
	utils.FailOnErrorWithMessage(err, "failed to setup SteampipeHub")

	// write a signature after everything gets done!
	// so that we can check for this later on
	writeDownloadedBinarySignature(dbImageDigest, fdwDigest)
}

func installSteampipeHub() error {
	StartService(InstallerInvoker)
	rawClient, err := createDbClient()
	if err != nil {
		return err
	}

	defer func() {
		rawClient.Close()
		// force stop
		StopDB(true)
	}()

	statements := []string{
		`DROP EXTENSION IF EXISTS "steampipe_postgres_fdw" CASCADE`,
		`CREATE EXTENSION IF NOT EXISTS "steampipe_postgres_fdw"`,
		`CREATE SERVER "steampipe" FOREIGN DATA WRAPPER "steampipe_postgres_fdw"`,
	}

	for _, statement := range statements {
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

package db

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/utils"
)

// StartService :: invokes `steampipe service start --database-listen local --refresh=false --invoker query`
func StartService(invoker Invoker) {
	utils.LogTime("db.StartService start")
	defer utils.LogTime("db.StartService end")
	log.Println("[TRACE] start service")
	// spawn a process to start the service, passing refresh=false to ensure we DO NOT refresh connections
	// (as we will do that ourselves)
	cmd := exec.Command(os.Args[0], "service", "start", "--database-listen", "local", "--refresh=false", "--invoker", string(invoker), "--install-dir", constants.SteampipeDir)
	startedAt := time.Now()
	spinnerShown := false
	startedChannel := make(chan bool, 1)
	errorChannel := make(chan error, 1)

	go func() {
		utils.LogTime("Starting CMD start")
		out, err := cmd.CombinedOutput()
		utils.LogTime("Starting CMD end")
		// we need to ignore errors when the invoker is the Installer
		// since when the installer starts the service, it will not be a stable state
		if err != nil && invoker != InvokerInstaller {
			errorChannel <- fmt.Errorf("Could not start steampipe service: %s", string(out))
			return
		}
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
				if cmdconfig.Viper().GetBool(constants.ConfigKeyShowInteractiveOutput) {
					s := display.ShowSpinner("Waiting for database to start...")
					defer display.StopSpinner(s)
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
	case <-startedChannel:
	// do nothing
	case x := <-errorChannel:
		utils.FailOnError(handleStartFailure(x))
	}
}

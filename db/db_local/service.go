package db_local

import (
	"errors"
	"log"
	"os"

	"github.com/turbot/steampipe/constants"

	"github.com/turbot/steampipe/utils"
)

// EnsureDbAndStartService ensures db is installed and starts service if necessary
func EnsureDbAndStartService(invoker constants.Invoker) error {
	utils.LogTime("db.EnsureDbAndStartService start")
	defer utils.LogTime("db.EnsureDbAndStartService end")

	log.Println("[TRACE] db.EnsureDbAndStartService start")

	if err := EnsureDBInstalled(); err != nil {
		return err
	}

	status, err := GetStatus()
	if err != nil {
		return errors.New("could not retrieve service status")
	}

	if status == nil {
		// the db service is not started - start it
		utils.LogTime("StartImplicitService start")
		log.Println("[TRACE] start implicit service")

		if _, err := StartDB(constants.DatabaseDefaultPort, ListenTypeLocal, invoker); err != nil {
			return err
		}
		utils.LogTime("StartImplicitService end")
	} else {
		// so db is already running - ensure it contains command schema
		// this is to handle the upgrade edge case where a user has a service running of an earlier version of steampipe
		// and upgrades to this version - we need to ensure we create the command schema
		return ensureCommandSchema()
	}
	return nil
}

// GetStatus checks that the db instance is running and returns its details
func GetStatus() (*RunningDBInstanceInfo, error) {
	utils.LogTime("db.GetStatus start")
	defer utils.LogTime("db.GetStatus end")

	info, err := loadRunningInstanceInfo()
	if err != nil {
		return nil, err
	}

	if info == nil {
		log.Println("[TRACE] GetRunStatus - loadRunningInstanceInfo returned nil ")
		// we do not have a info file
		return nil, nil
	}

	pidExists, err := PidExists(info.Pid)
	if err != nil {
		return nil, err
	}
	if !pidExists {
		log.Printf("[TRACE] GetRunStatus - pid %v does not exist\n", info.Pid)
		// nothing to do here
		os.Remove(runningInfoFilePath())
		return nil, nil
	}

	return info, nil
}

package db

import (
	"log"
	"os"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/turbot/steampipe/utils"
)

// GetStatus :: check that the db instance is running and returns it's details
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

	pidExists, err := psutils.PidExists(int32(info.Pid))
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

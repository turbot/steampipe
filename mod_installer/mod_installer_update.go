package mod_installer

import (
	"fmt"
)

// we are performing an update - verify that we have a lock file and andy specific mods requested for update
// exist in the lock file
func (i *ModInstaller) verifyCanUpdate() error {
	// TODO encapsulate this into lock and use from workspace load as well
	if len(i.installData.Lock.InstallCache) == 0 {
		return fmt.Errorf("no installation cache found - run 'steampipe plugin install'")
	}
	if len(i.installData.Lock.MissingVersions) > 0 {
		return fmt.Errorf("installation cache out of sync with installed mods - run 'steampipe plugin install'")
	}

	return nil
}

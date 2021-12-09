package mod_installer

import (
	"fmt"
	"log"
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

func (i *ModInstaller) shouldUpdate(modName string) bool {
	log.Printf("[TRACE] ModInstaller shouldUpdate %s", modName)
	if !i.updating {
		log.Printf("[TRACE] updates not enabled - returning false")
		return false
	}
	if len(i.UpdateMods) == 0 {
		log.Printf("[TRACE] no specific updates specified - returning true")
		return true
	}
	if i.UpdateMods[modName] {
		log.Printf("[TRACE] mod %s has been specified for update - returning true", modName)
		return true
	}
	log.Printf("[TRACE] mod %s has NOT been specified for update - returning true", modName)
	return false
}

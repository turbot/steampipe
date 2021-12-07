package mod_installer

import (
	"fmt"
	"log"
	"strings"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// we are performing an update - verify that we have a lock file and andy specific mods requested for update
// exist in the lock file
func (i *ModInstaller) verifyUpdates(updateMods []*modconfig.ModVersionConstraint) error {
	if len(i.installData.Lock) == 0 {
		return fmt.Errorf("no workspace lock file found - first run 'steampipe plugin install'")
	}
	i.UpdateMods = make(map[string]bool)

	// check all mods which have been requested to be updated exist in the lock file (ignore version)
	var missingMods []string
	for _, m := range updateMods {
		if i.installData.Lock.ContainsMod(m.Name) {
			// if this exists in the workspace lock, add to our map of updates
			i.UpdateMods[m.Name] = true
		} else {
			missingMods = append(missingMods, m.Name)
		}
	}
	if len(missingMods) != 0 {
		return fmt.Errorf("cannot update mod which is not a workspace dependency: %s", strings.Join(missingMods, ","))
	}
	return nil
}

func (i *ModInstaller) shouldUpdate(modName string) bool {
	log.Printf("[TRACE] ModInstaller shouldUpdate %s", modName)
	if !i.updateDependencies {
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

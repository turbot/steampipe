package migrate

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/turbot/steampipe/utils"
)

type Migrateable interface {
	MigrateFrom(old interface{}) Migrateable
	IsValid() bool
	Save() ([]byte, error)
}

func Migrate[O any, T Migrateable](old O, new T, oldPath string) error {
	stateFileContent, err := os.ReadFile(oldPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// deserialize into new
	err = json.Unmarshal(stateFileContent, &new)
	if err != nil {
		return err
	}

	// check whether we successfully derserialized into the new struct
	if new.IsValid() {
		return nil
	}

	// if schemaVersion is not available, deserialize into old
	err = json.Unmarshal(stateFileContent, &old)
	if err != nil {
		return err
	}
	// save the old file as a backup
	data, err := json.MarshalIndent(old, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(fmt.Sprintf("%s.backup", oldPath), data, 0644)
	if err != nil {
		return err
	}

	x := new.MigrateFrom(old)
	data, err = x.Save()
	if err != nil {
		return err
	}
	return utils.CombineErrors(os.Remove(oldPath), os.WriteFile(fmt.Sprintf("%s.migrated", oldPath), data, 0644))
}

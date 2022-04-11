package migrate

import (
	"encoding/json"
	"os"

	"github.com/turbot/steampipe/utils"
)

type Migrateable interface {
	MigrateFrom(old interface{}) Migrateable
	IsValid() bool
	Save() error
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

	// check for schemaVersion
	if new.IsValid() {
		return nil
	}

	// if schemaVersion is not available, deserialize into old
	err = json.Unmarshal(stateFileContent, &old)
	if err != nil {
		return err
	}
	x := new.MigrateFrom(old)
	return utils.CombineErrors(os.Remove(oldPath), x.Save())
}

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
	WriteOut() error
}

func Migrate(old interface{}, new Migrateable, oldPath string) error {
	stateFileContent, err := os.ReadFile(oldPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		fmt.Println("Could not read file")
		return err
	}

	// / deserialize into new
	err = json.Unmarshal(stateFileContent, new)
	if err != nil {
		fmt.Println("Could not parse file")
		return err
	}

	// check for schemaVersion
	if new.IsValid() {
		return nil
	}

	// if schemaVersion is not available, deserialize into old
	err = json.Unmarshal(stateFileContent, old)
	if err != nil {
		fmt.Println("Could not parse file")
		return err
	}
	new = new.MigrateFrom(old)
	return utils.CombineErrors(os.Remove(oldPath), new.WriteOut())
}

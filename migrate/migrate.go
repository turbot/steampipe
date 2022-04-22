package migrate

import (
	"encoding/json"
	"os"

	"github.com/turbot/steampipe/utils"
)

type Migrateable interface {
	MigrateFrom() Migrateable
	IsValid() bool
	Save() error
}

func Migrate(old Migrateable, oldPath string) error {
	stateFileContent, err := os.ReadFile(oldPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	// Deserialize into old struct
	err = json.Unmarshal(stateFileContent, &old)
	if err != nil {
		return err
	}

	// check whether we successfully derserialized into the new struct
	if old.IsValid() {
		return nil
	}

	x := old.MigrateFrom()
	return utils.CombineErrors(os.Remove(oldPath), x.Save())
}

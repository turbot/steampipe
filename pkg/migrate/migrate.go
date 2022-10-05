package migrate

import (
	"encoding/json"
	"os"

	"github.com/turbot/steampipe/pkg/error_helpers"
)

type Migrateable interface {
	MigrateFrom() Migrateable
	IsValid() bool
	Save() error
}

func Migrate(migrateable Migrateable, oldPath string) error {
	stateFileContent, err := os.ReadFile(oldPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	// Deserialize into old struct
	err = json.Unmarshal(stateFileContent, &migrateable)
	if err != nil {
		return err
	}

	// check whether we successfully derserialized into the new struct
	if migrateable.IsValid() {
		return nil
	}

	x := migrateable.MigrateFrom()
	return error_helpers.CombineErrors(os.Remove(oldPath), x.Save())
}

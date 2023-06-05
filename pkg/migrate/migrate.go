package migrate

import (
	"encoding/json"
	"log"
	"os"
)

type Migrateable interface {
	MigrateFrom() Migrateable
	IsValid() bool
	Save() error
}

func Migrate(migrateable Migrateable, oldPath string) {
	stateFileContent, err := os.ReadFile(oldPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("[INFO]", "nothing to migrate in", oldPath)
			return
		}
		log.Println("[ERROR]", "could not read file for migration:", oldPath, err)
	}
	// Deserialize into old struct
	err = json.Unmarshal(stateFileContent, &migrateable)
	if err != nil {
		log.Println("[ERROR]", "parsing failed for:", oldPath, err)
		return
	}

	// check whether we successfully derserialized into the new struct
	if migrateable.IsValid() {
		return
	}

	x := migrateable.MigrateFrom()

	if err := os.Remove(oldPath); err != nil {
		log.Println("[ERROR]", "could not remove after migration:", oldPath, err)
	}

	if err := x.Save(); err != nil {
		log.Println("[ERROR]", "could not save migrated data:", oldPath, err)
	}

}

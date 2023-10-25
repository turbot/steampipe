package db_local

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	filehelpers "github.com/turbot/go-kit/files"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
)

func TestTrimBackups(t *testing.T) {
	filepaths_steampipe.SteampipeDir, _ = filehelpers.Tildefy("~/.steampipe")
	// create backups more than MaxBackups
	backupDir := filepaths_steampipe.EnsureBackupsDir()
	filesCreated := []string{}
	for i := 0; i < constants_steampipe.MaxBackups; i++ {
		// make sure the files that get created end up to really old
		// this way we won't end up deleting any actual backup files
		timeLastYear := time.Now().Add(12 * 30 * 24 * time.Hour)

		fileName := fmt.Sprintf("database-%s-%2d", timeLastYear.Format("2006-01-02-15-04"), i)
		createFile := filepath.Join(backupDir, fileName)
		if err := os.WriteFile(filepath.Join(backupDir, fileName), []byte(""), 0644); err != nil {
			filesCreated = append(filesCreated, createFile)
		}
	}

	trimBackups()

	for _, f := range filesCreated {
		if filehelpers.FileExists(f) {
			t.Errorf("did not remove test backup file: %s", f)
		}
	}

}

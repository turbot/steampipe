package plugin

import (
	"log"
	"os"
	"time"

	"github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/filepaths"
)

func CleanupOldTmpDirs() {
	const tmpDirAgeThreshold = 24 * time.Hour
	tmpDirs, err := files.ListFiles(filepaths_steampipe.EnsurePluginDir(), &files.ListOptions{
		Include: []string{"tmp-*"},
		Flags:   files.DirectoriesRecursive,
	})
	if err != nil {
		log.Printf("[TRACE] Error while globbing for tmp dirs in plugin dir: %s", err)
		return
	}

	for _, tmpDir := range tmpDirs {
		stat, err := os.Stat(tmpDir)
		if err != nil {
			log.Printf("[TRACE] Error while stating tmp dir %s: %s", tmpDir, err)
			continue
		}
		if time.Since(stat.ModTime()) > tmpDirAgeThreshold {
			if err := os.RemoveAll(tmpDir); err != nil {
				log.Printf("[TRACE] Error while removing old tmp dir %s: %s", tmpDir, err)
			}
		}
	}
}

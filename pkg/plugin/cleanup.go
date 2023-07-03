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
	tmpDirs, err := files.ListFiles(filepaths.EnsurePluginDir(), &files.ListOptions{
		Include: []string{"tmp-*"},
		Flags:   files.DirectoriesRecursive,
	})
	if err != nil {
		log.Printf("Error while globbing for tmp dirs in plugin dir: %s\n", err)
		return
	}

	for _, tmpDir := range tmpDirs {
		stat, err := os.Stat(tmpDir)
		if err != nil {
			log.Printf("Error while stating tmp dir %s: %s\n", tmpDir, err)
			continue
		}
		if time.Since(stat.ModTime()) > tmpDirAgeThreshold {
			if err := os.RemoveAll(tmpDir); err != nil {
				log.Printf("Error while removing old tmp dir %s: %s\n", tmpDir, err)
			}
		}
	}
}

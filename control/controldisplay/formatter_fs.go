package controldisplay

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/turbot/steampipe/filepaths"
)

//go:embed templates/*
var builtinTemplateFS embed.FS

// EnsureTemplates scans the '$STEAMPIPE_INSTALL_DIR/templates' directory and
// copies over any missing templates as defined in the 'templates' package
//
// The name of the folder in the 'templates' package is used to identify
// templates in '$STEAMPIPE_INSTALL_DIR/templates' - where it is expected
// that a directory with the same name will exist. If said directory does
// not exist, it is copied over from 'templates'
//
func EnsureTemplates() error {
	log.Println("[TRACE] ensuring check export/output templates")
	dirs, err := fs.ReadDir(builtinTemplateFS, "templates")
	if err != nil {
		return err
	}
	for _, d := range dirs {
		targetDirectory := filepath.Join(filepaths.EnsureTemplateDir(), d.Name())
		if _, err := os.Stat(targetDirectory); os.IsNotExist(err) {
			log.Println("[TRACE] target directory does not exist - copying template")
			if err := writeTemplate(d.Name(), targetDirectory); err != nil {
				log.Println("[ERROR] error copying template", err)
				return err
			}
		} else if err != nil {
			log.Println("[ERROR] error fetching directory information", err)
			return err
		}
	}
	return nil
}

func writeTemplate(path string, target string) error {
	err := os.MkdirAll(target, 0744)
	if err != nil {
		return err
	}

	entries, err := fs.ReadDir(builtinTemplateFS, filepath.Join("templates", path))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		sourceInEmbedFs := filepath.Join("templates", path, entry.Name())
		bytes, err := fs.ReadFile(builtinTemplateFS, sourceInEmbedFs)
		if err != nil {
			return err
		}

		err = os.WriteFile(filepath.Join(target, entry.Name()), bytes, 0744)
		if err != nil {
			return err
		}
	}

	return nil
}

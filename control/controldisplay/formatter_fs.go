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
		log.Println("[ERROR] error reading embedded filesystem", err)
		return err
	}
	for _, d := range dirs {
		log.Println("[TRACE] embedded template:", d)
		targetDirectory := filepath.Join(filepaths.TemplateDir(), d.Name())
		log.Println("[TRACE] target directory:", targetDirectory)
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
	log.Println("[TRACE] creating directory", target)
	err := os.MkdirAll(target, 0744)
	if err != nil {
		log.Println("[ERROR] error creating template directory", err)
		return err
	}

	log.Println("[TRACE] reading entries from embedded FS")
	entries, err := fs.ReadDir(builtinTemplateFS, filepath.Join("templates", path))
	if err != nil {
		log.Println("[ERROR] error reading entries from embedded FS", err)
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			log.Println("[TRACE] found directory in template - skipping")
			continue
		}
		sourceInEmbedFs := filepath.Join("templates", path, entry.Name())
		log.Println("[TRACE] reading source template file", sourceInEmbedFs)
		bytes, err := fs.ReadFile(builtinTemplateFS, sourceInEmbedFs)
		if err != nil {
			log.Println("[ERROR] reading source template file", err)
			return err
		}

		log.Println("[TRACE] Writing to", filepath.Join(target, entry.Name()))
		err = os.WriteFile(filepath.Join(target, entry.Name()), bytes, 0744)
		if err != nil {
			log.Println("[ERROR] writing to target file", err)
			return err
		}
	}

	return nil
}

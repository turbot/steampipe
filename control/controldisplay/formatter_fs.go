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

// WriteTemplates scans the '$STEAMPIPE_INSTALL_DIR/templates' directory and
// copies over all the templates defined in the 'templates' package
//
// The name of the folder in the 'templates' package is used to identify
// templates in '$STEAMPIPE_INSTALL_DIR/templates' - where it is expected
// that a directory with the same name will exist.
//
// We always re-write the templates(even if they exist or not) so that the
// latest changes in the 'templates' package are always reflected in
// '$STEAMPIPE_INSTALL_DIR/templates'.
func WriteTemplates() error {
	log.Println("[TRACE] writing check export/output templates")
	dirs, err := fs.ReadDir(builtinTemplateFS, "templates")
	if err != nil {
		return err
	}
	for _, d := range dirs {
		targetDirectory := filepath.Join(filepaths.EnsureTemplateDir(), d.Name())
		// write the templates for each directory
		if err := writeTemplate(d.Name(), targetDirectory); err != nil {
			log.Println("[ERROR] error copying template", err)
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

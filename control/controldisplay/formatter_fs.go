package controldisplay

import (
	"embed"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/turbot/steampipe/filepaths"
)

//go:embed templates/*
var builtinTemplateFS embed.FS

type TemplateVersionFile struct {
	Version string `json:"version"`
}

// WriteTemplates scans the '$STEAMPIPE_INSTALL_DIR/check/templates' directory and
// copies over the templates defined in the 'templates' package if needed.
//
// The name of the folder in the 'templates' package is used to identify
// templates in '$STEAMPIPE_INSTALL_DIR/templates' - where it is expected
// that a directory with the same name will exist.
//
// We re-write the templates, when there is a higher template version
// available in the 'templates' package.
func WriteTemplates() error {
	start := time.Now()
	log.Println("[TRACE] ensuring check export/output templates")
	dirs, err := fs.ReadDir(builtinTemplateFS, "templates")
	if err != nil {
		return err
	}
	for _, d := range dirs {
		targetDirectory := filepath.Join(filepaths.EnsureTemplateDir(), d.Name())
		currentVersionsFilePath := filepath.Join(targetDirectory, "versions.json")
		embeddedVersionsFilePath := filepath.Join("templates", d.Name(), "versions.json")

		// check if version in versions.json matches with embedded template version
		if getCurrentTemplateVersion(currentVersionsFilePath) != getEmbeddedTemplateVersion(embeddedVersionsFilePath) {
			log.Println("[TRACE] versions do not match - copying template")
			if err := writeTemplate(d.Name(), targetDirectory); err != nil {
				log.Println("[TRACE] error copying template", err)
				return err
			}
		}
	}
	elapsed := time.Since(start)
	log.Printf("[WARN] >> time elapsed: %v", elapsed)
	return nil
}

func getCurrentTemplateVersion(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("[TRACE] template version file does not exist - install the new template")
		} else {
			log.Println("[TRACE] error reading current version file - installing the new template")
		}
		return ""
	}
	var ver TemplateVersionFile
	err = json.Unmarshal(data, &ver)
	if err != nil {
		log.Println("[TRACE] error while unmarshaling current versions.json file", err)
		return ""
	}
	return ver.Version
}

func getEmbeddedTemplateVersion(path string) string {
	data, err := fs.ReadFile(builtinTemplateFS, path)
	if err != nil {
		log.Println("[TRACE] error reading embedded version file - installing the new template")
		return ""
	}
	var ver TemplateVersionFile
	err = json.Unmarshal(data, &ver)
	if err != nil {
		log.Println("[TRACE] error while unmarshaling json", err)
		return ""
	}
	return ver.Version
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

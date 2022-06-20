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
// copies over all the templates defined in the 'templates' package if needed.
//
// The name of the folder in the 'templates' package is used to identify
// templates in '$STEAMPIPE_INSTALL_DIR/templates' - where it is expected
// that a directory with the same name will exist.
//
// We only re-write the templates, when there is a higher template version
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
		targetVersionsFilePath := filepath.Join(targetDirectory, "versions.json")
		embeddedVersionsFilePath := filepath.Join("templates", d.Name(), "versions.json")

		// check if versions file exist
		if _, err := os.Stat(targetVersionsFilePath); os.IsNotExist(err) {
			log.Println("[TRACE] target versions file does not exist - copying template")
			if err := writeTemplate(d.Name(), targetDirectory); err != nil {
				log.Println("[ERROR] error copying template", err)
				return err
			}
		} else if err != nil {
			log.Println("[ERROR] error fetching directory information", err)
			return err
		}

		// now check if version in versions.json matches with current template version
		if getTargetTemplateVersion(targetVersionsFilePath) != getEmbedTemplateVersion(embeddedVersionsFilePath) {
			log.Println("[TRACE] target versions do not match - copying template")
			if err := writeTemplate(d.Name(), targetDirectory); err != nil {
				log.Println("[ERROR] error copying template", err)
				return err
			}
		}
	}
	elapsed := time.Since(start)
	log.Printf("[WARN] >> time elapsed: %v", elapsed)
	return nil
}

func getTargetTemplateVersion(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("[ERROR] error reading target version file", err)
		return ""
	}
	var ver TemplateVersionFile
	err = json.Unmarshal(data, &ver)
	if err != nil {
		log.Println("[ERROR] error while unmarshaling json", err)
		return ""
	}
	return ver.Version
}

func getEmbedTemplateVersion(path string) string {
	data, err := fs.ReadFile(builtinTemplateFS, path)
	if err != nil {
		log.Println("[ERROR] error reading embedded versionfile", err)
		return ""
	}
	var ver TemplateVersionFile
	err = json.Unmarshal(data, &ver)
	if err != nil {
		log.Println("[ERROR] error while unmarshaling json", err)
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

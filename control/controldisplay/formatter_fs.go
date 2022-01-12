package controldisplay

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/turbot/steampipe/filepaths"
)

//go:embed templates/*
var builtinTemplateFS embed.FS

func EnsureTemplates() error {
	dirs, err := fs.ReadDir(builtinTemplateFS, "templates")
	if err != nil {
		return err
	}
	for _, d := range dirs {
		targetDirectory := filepath.Join(filepaths.TemplateDir(), d.Name())
		if _, err := os.Stat(targetDirectory); os.IsNotExist(err) {
			if err := writeTemplate(d.Name(), targetDirectory); err != nil {
				return err
			}
		} else if err != nil {
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

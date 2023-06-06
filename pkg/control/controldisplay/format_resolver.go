package controldisplay

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/export"
	"github.com/turbot/steampipe/pkg/filepaths"
)

type FormatResolver struct {
	// array of unique formatters used for export
	exportFormatters []Formatter
	formatterByName  map[string]Formatter
}

func NewFormatResolver() (*FormatResolver, error) {
	templates, err := loadAvailableTemplates()
	if err != nil {
		return nil, err
	}

	formatters := []Formatter{
		&NullFormatter{},
		&TextFormatter{},
		&SnapshotFormatter{},
	}

	res := &FormatResolver{
		formatterByName: make(map[string]Formatter),
	}

	for _, f := range formatters {
		if err := res.registerFormatter(f); err != nil {
			return nil, err
		}
	}
	for _, t := range templates {
		f, err := NewTemplateFormatter(t)
		if err != nil {
			return nil, err
		}

		if err := res.registerFormatter(f); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (r *FormatResolver) GetFormatter(arg string) (Formatter, error) {
	if formatter, found := r.formatterByName[arg]; found {
		return formatter, nil
	}

	return nil, fmt.Errorf(" invalid output format: '%s'", arg)
}

func (r *FormatResolver) registerFormatter(f Formatter) error {
	name := f.Name()

	if _, ok := r.formatterByName[name]; ok {
		return fmt.Errorf("failed to register output formatter - duplicate format name %s", name)
	}
	r.formatterByName[name] = f
	// if the formatter has an alias, also register by alias
	if alias := f.Alias(); alias != "" {
		if _, ok := r.formatterByName[alias]; ok {
			return fmt.Errorf("failed to register output formatter - duplicate format name %s", alias)
		}
		r.formatterByName[alias] = f
	}
	// add to exportFormatters list (exclude 'None')
	if f.Name() != constants.OutputFormatNone {
		r.exportFormatters = append(r.exportFormatters, f)
	}
	return nil
}

func (r *FormatResolver) controlExporters() []export.Exporter {
	res := make([]export.Exporter, len(r.exportFormatters))
	for i, formatter := range r.exportFormatters {
		res[i] = NewControlExporter(formatter)

	}
	return res
}

func loadAvailableTemplates() ([]*OutputTemplate, error) {
	templateDir := filepaths.EnsureTemplateDir()
	templateDirectories, err := os.ReadDir(templateDir)
	if err != nil {
		return nil, err
	}
	templates := []*OutputTemplate{}
	for _, f := range templateDirectories {
		if shouldIgnoreDirectory(templateDir, f) {
			continue
		}
		templates = append(templates, NewOutputTemplate(filepath.Join(templateDir, f.Name())))
	}

	return templates, nil
}

var directoryIgnoreList []string = []string{
	".DS_Store",
}

func shouldIgnoreDirectory(templateDir string, dir fs.DirEntry) bool {
	if !dir.IsDir() {
		return true
	}
	if helpers.StringSliceContains(directoryIgnoreList, dir.Name()) {
		return true
	}
	dirLs, err := os.ReadDir(filepath.Join(templateDir, dir.Name()))
	if err != nil {
		// ignore this directory if this cannot be listed at least
		return true
	}
	if len(dirLs) == 0 {
		// ignore this directory if there are no files in it.
		return true
	}
	return false
}

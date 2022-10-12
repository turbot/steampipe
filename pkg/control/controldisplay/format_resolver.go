package controldisplay

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/export"
	"github.com/turbot/steampipe/pkg/filepaths"
	"os"
	"path/filepath"
)

type FormatResolver struct {
	templates       []*OutputTemplate
	formatterByName map[string]Formatter
	// array of unique formatters
	formatters []Formatter
}

func NewFormatResolver() (*FormatResolver, error) {
	templates, err := loadAvailableTemplates()
	if err != nil {
		return nil, err
	}

	formatters := []Formatter{
		//&NullFormatter{},
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

	return nil, fmt.Errorf("could not resolve formatter for %s", arg)
}

func (r *FormatResolver) registerFormatter(f Formatter) error {
	name := f.Name()

	if _, ok := r.formatterByName[name]; ok {
		return fmt.Errorf("failed to register output formatter - duplicate format name %s", name)
	}
	r.formatterByName[name] = f
	if alias := f.Alias(); alias != "" {
		if _, ok := r.formatterByName[alias]; ok {
			return fmt.Errorf("failed to register output formatter - duplicate format name %s", alias)
		}
		r.formatterByName[alias] = f
	}
	// add to unique formatter list
	r.formatters = append(r.formatters, f)
	return nil
}

func (r *FormatResolver) controlExporters() []export.Exporter {
	res := make([]export.Exporter, len(r.formatters))
	for i, formatter := range r.formatters {
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
	templates := make([]*OutputTemplate, len(templateDirectories))
	for idx, f := range templateDirectories {
		templates[idx] = NewOutputTemplate(filepath.Join(templateDir, f.Name()))
	}

	return templates, nil
}

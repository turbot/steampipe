package controldisplay

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/export"
	"github.com/turbot/steampipe/pkg/filepaths"
	"os"
	"path"
	"path/filepath"
)

type FormatResolver struct {
	templates       []*OutputTemplate
	formatterByName map[string]Formatter

	formatterByExtension map[string]Formatter
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
		formatterByName:      make(map[string]Formatter),
		formatterByExtension: make(map[string]Formatter),
	}

	for _, f := range formatters {
		if err := res.registerFormatter(f); err != nil {
			return nil, err

		}
	}
	for _, t := range templates {
		if err := res.registerTemplate(t); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (r *FormatResolver) registerFormatter(f Formatter) error {
	ext := f.FileExtension()
	name := f.Name()
	if _, ok := r.formatterByExtension[ext]; ok {
		return fmt.Errorf("failed to register output formatter - duplicate extension %s", ext)
	}
	r.formatterByExtension[ext] = f

	if _, ok := r.formatterByName[name]; ok {
		return fmt.Errorf("failed to register output formatter - duplicate format name %s", name)
	}
	r.formatterByName[name] = f
	return nil
}

func (r *FormatResolver) registerTemplate(t *OutputTemplate) error {
	f, err := NewTemplateFormatter(t)
	if err != nil {
		return err
	}

	if _, ok := r.formatterByName[t.FormatName]; ok {
		return fmt.Errorf("failed to register output template - duplicate format name %s", t.FormatName)
	}
	r.formatterByName[t.FormatName] = f

	if _, ok := r.formatterByName[t.FormatFullName]; ok {
		return fmt.Errorf("failed to register output template - duplicate format name %s", t.FormatFullName)
	}
	r.formatterByName[t.FormatFullName] = f

	// now register extension
	if existing, ok := r.formatterByExtension[t.FileExtension]; ok {
		existingIsDefaultForExt := existing.(*TemplateFormatter).exportFormat.DefaultTemplateForExtension
		newIsDefaultForExt := t.DefaultTemplateForExtension

		// check if either the existing or new template is the default for extension
		if newIsDefaultForExt && existingIsDefaultForExt ||
			!newIsDefaultForExt && !existingIsDefaultForExt {
			// both or neither are default for the extension - this is an error
			return fmt.Errorf("failed to register output template - duplicate extension %s", t.FileExtension)
		}

		if existingIsDefaultForExt {
			// if existing is default and new isn't, nothing to do
			return nil
		}
	}
	r.formatterByExtension[t.FileExtension] = f

	return nil
}

func (r *FormatResolver) GetFormatter(arg string) (Formatter, error) {

	if formatter, found := r.formatterByName[arg]; found {
		return formatter, nil
	}
	if formatter, found := r.formatterByExtension[path.Ext(arg)]; found {
		return formatter, nil
	}

	return nil, fmt.Errorf("could not resolve formatter for %s", arg)
}

func (r *FormatResolver) controlExporters() (exportersByName, exportersByExtension map[string]export.Exporter) {
	exportersByName = make(map[string]export.Exporter)
	exportersByExtension = make(map[string]export.Exporter)
	allExporters := make(map[Formatter]export.Exporter)

	for _, formatter := range r.formatterByName {
		if _, ok := allExporters[formatter]; !ok {
			allExporters[formatter] = NewControlExporter(formatter)
		}
	}

	// now build the name and extension map using the map of exportes as the source
	for name, formatter := range r.formatterByName {
		exporter := allExporters[formatter]
		exportersByName[name] = exporter
	}
	for name, formatter := range r.formatterByExtension {
		exporter := allExporters[formatter]
		exportersByExtension[name] = exporter
	}
	return
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

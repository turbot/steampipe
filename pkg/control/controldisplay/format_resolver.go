package controldisplay

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"os"
	"path/filepath"
	"strings"
)

type FormatResolver struct {
	templates        []*OutputTemplate
	outputFormatters map[string]Formatter
}

func NewFormatResolver() (*FormatResolver, error) {
	templates, err := loadAvailableTemplates()
	if err != nil {
		return nil, err
	}
	var outputFormatters = map[string]Formatter{
		constants.OutputFormatNone:     &NullFormatter{},
		constants.OutputFormatText:     &TextFormatter{},
		constants.OutputFormatBrief:    &TextFormatter{},
		constants.OutputFormatSnapshot: &SnapshotFormatter{},
	}

	return &FormatResolver{templates: templates, outputFormatters: outputFormatters}, nil
}

func (r *FormatResolver) GetFormatter(arg string) (Formatter, error) {
	if formatter, found := r.outputFormatters[arg]; found {
		return formatter, nil
	}

	// otherwise look for a template
	templateFormat, err := r.resolveOutputTemplate(arg)
	if err != nil {
		return nil, err
	}
	return NewTemplateFormatter(templateFormat)
}

func (r *FormatResolver) GetFormatterByExtension(filename string) (Formatter, error) {
	// so we failed to exactly match an existing format or template name
	// instead, treat the arg as a filename and try to infer the template from the extension
	// try to find the target template by the given filename
	matchedTemplate, err := r.findTemplateByExtension(filename)
	if err != nil {
		return nil, err
	}
	return NewTemplateFormatter(matchedTemplate)
}

// resolveOutputTemplate accepts the export argument and resolves the template to use.
// If an exact match to the available templates is not found, and if 'allowFilenameEvaluation' is true
// then the 'export' value is parsed as a filename and the suffix is used to match to available templates
// returns
// - the export template to use
// - the path of the file to write to
// - error (if any)
func (r *FormatResolver) resolveOutputTemplate(export string) (format *OutputTemplate, err error) {
	// try an exact match
	for _, t := range r.templates {
		if t.FormatName == export || t.FormatFullName == export {
			return t, nil
		}
	}

	return nil, fmt.Errorf("template %s not found", export)
}

func (r *FormatResolver) findTemplateByExtension(filename string) (format *OutputTemplate, err error) {
	// does the filename end with this exact format?
	for _, t := range r.templates {
		if strings.HasSuffix(filename, t.FormatFullName) {
			return t, nil
		}
	}

	extension := filepath.Ext(filename)
	if len(extension) == 0 {
		// we don't have anything to work with
		return nil, fmt.Errorf("template %s not found", filename)
	}
	var matchingTemplates []*OutputTemplate

	// does the given extension match with one of the template extension?
	for _, t := range r.templates {
		if strings.HasSuffix(t.OutputExtension, extension) {
			matchingTemplates = append(matchingTemplates, t)
		}
	}

	if len(matchingTemplates) > 1 {
		matchNames := []string{}
		// find out if any of them has preference
		for _, match := range matchingTemplates {
			if match.DefaultTemplateForExtension {
				return match, nil
			}
			matchNames = append(matchNames, match.FormatName)
		}
		// there's ambiguity - we have more than one matching templates based on extension
		return nil, fmt.Errorf("ambiguous templates found: %v", matchNames)
	}

	if len(matchingTemplates) == 1 {
		return matchingTemplates[0], nil
	}

	return nil, fmt.Errorf("template %s not found", filename)
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

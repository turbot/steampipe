package controldisplay

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/turbot/steampipe/filepaths"
)

type ExportTemplate struct {
	TemplatePath                string
	FormatName                  string
	OutputExtension             string
	FormatFullName              string
	DefaultTemplateForExtension bool
}

func NewExportTemplate(directory string) *ExportTemplate {
	format := new(ExportTemplate)
	format.TemplatePath = directory

	directory = filepath.Base(directory)

	// try splitting by a .(dot)
	lastDotIndex := strings.LastIndex(directory, ".")
	if lastDotIndex == -1 {
		format.OutputExtension = fmt.Sprintf(".%s", directory)
		format.FormatName = directory
		format.DefaultTemplateForExtension = true
	} else {
		format.OutputExtension = filepath.Ext(directory)
		format.FormatName = strings.TrimSuffix(directory, filepath.Ext(directory))
	}
	format.FormatFullName = fmt.Sprintf("%s%s", format.FormatName, format.OutputExtension)
	return format
}

func (ft ExportTemplate) String() string {
	return fmt.Sprintf("( %s %s %s %s )", ft.TemplatePath, ft.FormatName, ft.OutputExtension, ft.FormatFullName)
}

// GetExportTemplate accepts the export argument and tries to figure out the template to use
// if an exact match to the available templates is not found, and if 'allowFilenameEvaluation' is true
// then the 'export' value is parsed as a filename and the suffix is used to match to available templates
// returns
// - the export template to use
// - the path of the file to write to
// - error (if any)
func GetExportTemplate(export string, allowFilenameEvaluation bool) (format *ExportTemplate, targetFilename string, err error) {
	available, err := loadAvailableTemplates()
	if err != nil {
		return nil, "", err
	}

	// try an exact match
	for _, t := range available {
		if t.FormatName == export || t.FormatFullName == export {
			return t, "", nil
		}
	}

	if !allowFilenameEvaluation {
		return nil, "", ErrTemplateNotFound
	}

	// if the above didn't match, then the input argument is a file name
	targetFilename = export

	// try to find the target template by the given filename
	matchedTemplate, err := findTemplateByFilename(targetFilename, available)

	return matchedTemplate, targetFilename, err
}

func findTemplateByFilename(export string, available []*ExportTemplate) (format *ExportTemplate, err error) {
	// does the filename end with this exact format?
	for _, t := range available {
		if strings.HasSuffix(export, t.FormatFullName) {
			return t, nil
		}
	}

	extension := filepath.Ext(export)
	if len(extension) == 0 {
		// we don't have anything to work with
		return nil, ErrTemplateNotFound
	}
	matchingTemplates := []*ExportTemplate{}

	// does the given extension match with one of the template extension?
	for _, t := range available {
		if strings.HasSuffix(t.OutputExtension, extension) {
			matchingTemplates = append(matchingTemplates, t)
		}
	}

	if len(matchingTemplates) > 1 {
		// find out if any of them has preference
		for _, match := range matchingTemplates {
			if match.DefaultTemplateForExtension {
				return match, nil
			}
		}
		// there's ambiguity - we have more than one matching templates based on extension
		return nil, ErrAmbiguousTemplate
	}

	if len(matchingTemplates) == 1 {
		return matchingTemplates[0], nil
	}

	return nil, ErrTemplateNotFound
}

func loadAvailableTemplates() ([]*ExportTemplate, error) {
	templateDir := filepaths.TemplateDir()
	templateDirectories, err := os.ReadDir(templateDir)
	if err != nil {
		return nil, err
	}
	templates := make([]*ExportTemplate, len(templateDirectories))
	for idx, f := range templateDirectories {
		templates[idx] = NewExportTemplate(filepath.Join(templateDir, f.Name()))
	}

	return templates, nil
}

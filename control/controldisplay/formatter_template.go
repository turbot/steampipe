package controldisplay

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/turbot/steampipe/control/controlexecute"
	"github.com/turbot/steampipe/filepaths"
)

var ErrAmbiguousTemplate = errors.New("ambiguous templates found")
var ErrTemplateNotFound = errors.New("template not found")

// TemplateFormatter implements the 'Formatter' interface and exposes a generic template based output mechanism
// for 'check' execution trees
type TemplateFormatter struct {
	template     *template.Template
	exportFormat ExportTemplate
}

func (tf TemplateFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	reader, writer := io.Pipe()
	go func() {
		if err := tf.template.ExecuteTemplate(writer, "outlet", tree); err != nil {
			writer.CloseWithError(err)
		} else {
			writer.Close()
		}
	}()
	return reader, nil
}

func (tf TemplateFormatter) FileExtension() string {
	// if the extension is the same as the format name, return just the extension
	if tf.exportFormat.DefaultTemplateForExtension {
		return tf.exportFormat.OutputExtension
	} else {
		// otherwise return the fullname
		return tf.exportFormat.FormatFullName
	}
}

func CreateTemplateFormatter(input ExportTemplate) (*TemplateFormatter, error) {
	t, err := template.New("outlet").
		Funcs(formatterTemplateFuncMap).
		ParseFS(os.DirFS(input.TemplatePath), "*")

	if err != nil {
		return nil, err
	}
	return &TemplateFormatter{exportFormat: input, template: t}, nil
}

// GetExportTemplate accepts the export argument and tries to figure out the template to use
// if an exact match to the available templates is not found, and if 'allowFilenameEvaluation' is true
// then the 'export' value is parsed as a filename and the suffix is used to match to available templates
func GetExportTemplate(export string, allowFilenameEvaluation bool) (format *ExportTemplate, filename string, err error) {
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
	filename = export

	// try to find the target template by the given filename
	matchedTemplate, err := findTemplateByFilename(filename, available)

	return matchedTemplate, filename, err
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

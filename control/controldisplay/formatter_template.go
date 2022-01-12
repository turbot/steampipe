package controldisplay

import (
	"context"
	"errors"
	"fmt"
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

type TemplateFormatter struct {
	outputExtension string
	template        *template.Template
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
	return tf.outputExtension
}

func CreateTemplateFormatter(input ExportTemplate) (*TemplateFormatter, error) {
	t, err := template.New("outlet").
		Funcs(formatterTemplateFuncMap).
		ParseFS(os.DirFS(input.TemplatePath), "*")

	if err != nil {
		return nil, err
	}
	return &TemplateFormatter{outputExtension: input.OutputExtension, template: t}, nil
}

func GetExportTemplate(export string) (format *ExportTemplate, filename string, err error) {
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

	// if the above didn't match, then the input argument is a file name
	filename = export

	// try to find the target template by the given filename
	matchedTemplate, err := findTemplateByFilename(filename, available)

	return matchedTemplate, filename, err
}

func findTemplateByFilename(export string, available []*ExportTemplate) (format *ExportTemplate, err error) {
	// does the export end with this exact format?
	for _, t := range available {
		if strings.HasSuffix(export, t.FormatFullName) {
			return t, nil
		}
	}

	extension := filepath.Ext(export)
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
		// there's ambiguity
		fmt.Println(matchingTemplates)
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
		templates[idx] = NewFormatTemplateFromDirectoryName(filepath.Join(templateDir, f.Name()))
	}

	return templates, nil
}

func NewFormatTemplateFromDirectoryName(directory string) *ExportTemplate {
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

type ExportTemplate struct {
	TemplatePath                string
	FormatName                  string
	OutputExtension             string
	FormatFullName              string
	DefaultTemplateForExtension bool
}

func (ft ExportTemplate) String() string {
	return fmt.Sprintf("( %s %s %s %s )", ft.TemplatePath, ft.FormatName, ft.OutputExtension, ft.FormatFullName)
}

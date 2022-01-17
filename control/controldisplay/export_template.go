package controldisplay

import (
	"fmt"
	"path/filepath"
	"strings"
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

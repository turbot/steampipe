package controldisplay

import (
	"fmt"
	"path/filepath"
	"strings"
)

type OutputTemplate struct {
	TemplatePath                string
	FormatName                  string
	FileExtension               string
	FormatFullName              string
	DefaultTemplateForExtension bool
}

func NewOutputTemplate(directory string) *OutputTemplate {
	format := new(OutputTemplate)
	format.TemplatePath = directory

	directory = filepath.Base(directory)

	// try splitting by a .(dot)
	lastDotIndex := strings.LastIndex(directory, ".")
	if lastDotIndex == -1 {
		format.FileExtension = fmt.Sprintf(".%s", directory)
		format.FormatName = directory
		format.DefaultTemplateForExtension = true
	} else {
		format.FileExtension = filepath.Ext(directory)
		format.FormatName = strings.TrimSuffix(directory, filepath.Ext(directory))
	}
	format.FormatFullName = fmt.Sprintf("%s%s", format.FormatName, format.FileExtension)
	return format
}

func (ft *OutputTemplate) String() string {
	return fmt.Sprintf("( %s %s %s %s )", ft.TemplatePath, ft.FormatName, ft.FileExtension, ft.FormatFullName)
}

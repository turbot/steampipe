package controldisplay

import (
	"fmt"
	"path/filepath"
	"strings"
)

type OutputTemplate struct {
	TemplatePath   string
	FormatName     string
	FileExtension  string
	FormatFullName string
}

func NewOutputTemplate(directoryPath string) *OutputTemplate {
	format := new(OutputTemplate)
	format.TemplatePath = directoryPath

	directoryName := filepath.Base(directoryPath)
	// does the directory name include an extension?
	ext := filepath.Ext(directoryName)
	format.FormatFullName = directoryName
	format.FormatName = strings.TrimSuffix(directoryName, ext)
	format.FileExtension = fmt.Sprintf(".%s", directoryName)

	return format
}

func (ft *OutputTemplate) String() string {
	return fmt.Sprintf("( %s %s %s %s )", ft.TemplatePath, ft.FormatName, ft.FileExtension, ft.FormatFullName)
}

package controldisplay

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/control/controlexecute"
)

var ErrFormatterNotFound = errors.New("Formatter not found")

type FormatterMap map[string]Formatter

var outputFormatters FormatterMap = FormatterMap{
	constants.CheckOutputFormatNone:  &NullFormatter{},
	constants.CheckOutputFormatText:  &TextFormatter{},
	constants.CheckOutputFormatBrief: &TextFormatter{},
}

type CheckExportTarget struct {
	Formatter Formatter
	File      string
}

func NewCheckExportTarget(formatter Formatter, file string) CheckExportTarget {
	return CheckExportTarget{
		Formatter: formatter,
		File:      file,
	}
}

type Formatter interface {
	Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error)
	FileExtension() string
}

func GetTemplateExportFormatter(arg string, allowFilenameEvaluation bool) (Formatter, string, error) {
	templateFormat, fileName, err := ResolveExportTemplate(arg, allowFilenameEvaluation)
	if err != nil {
		return nil, "", err
	}
	formatter, err := NewTemplateFormatter(*templateFormat)
	return formatter, fileName, err
}

func GetDefinedOutputFormatter(outputFormat string) (Formatter, bool) {
	formatter, found := outputFormatters[outputFormat]
	return formatter, found
}

// NullFormatter is to be used when no output is expected. It always returns a `io.Reader` which
// reads an empty string
type NullFormatter struct{}

func (j *NullFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	return strings.NewReader(""), nil
}

func (j *NullFormatter) FileExtension() string {
	// will not be called
	return ""
}

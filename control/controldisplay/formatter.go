package controldisplay

import (
	"context"
	"errors"
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/MasterMinds/sprig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/control/controlexecute"
)

var ErrFormatterNotFound = errors.New("Formatter not found")

type FormatterMap map[string]Formatter

var outputFormatters FormatterMap = FormatterMap{
	constants.CheckOutputFormatNone:  &NullFormatter{},
	constants.CheckOutputFormatCSV:   &CSVFormatter{},
	constants.CheckOutputFormatText:  &TextFormatter{},
	constants.CheckOutputFormatBrief: &TextFormatter{},
}

var exportFormatters FormatterMap = FormatterMap{
	constants.CheckOutputFormatCSV: &CSVFormatter{},
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

func GetDefinedExportFormatter(arg string) (Formatter, bool) {
	formatter, found := exportFormatters[arg]
	return formatter, found
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

func templateFuncs() template.FuncMap {
	useFromSprigMap := []string{"upper", "toJson", "quote", "dict", "add", "now", "toPrettyJson"}

	var funcs template.FuncMap = template.FuncMap{}
	sprigMap := sprig.TxtFuncMap()
	for _, use := range useFromSprigMap {
		f, found := sprigMap[use]
		if found {
			funcs[use] = f
		}
	}
	for k, v := range formatterTemplateFuncMap {
		funcs[k] = v
	}

	return funcs
}

var formatterTemplateFuncMap template.FuncMap = template.FuncMap{
	"durationInSeconds": func(t time.Duration) float64 { return t.Seconds() },
}

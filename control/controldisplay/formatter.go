package controldisplay

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/MasterMinds/sprig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/control/controlexecute"
	"github.com/turbot/steampipe/version"
)

var ErrFormatterNotFound = errors.New("Formatter not found")

type FormatterMap map[string]Formatter

func (m FormatterMap) keys() []string {
	keys := make([]string, len(m))
	i := 0
	for key := range m {
		keys[i] = key
		i++
	}
	return keys
}

var outputFormatters FormatterMap = FormatterMap{
	constants.CheckOutputFormatNone:  &NullFormatter{},
	constants.CheckOutputFormatJSON:  &JSONFormatter{},
	constants.CheckOutputFormatText:  &TextFormatter{},
	constants.CheckOutputFormatBrief: &TextFormatter{},
}

var exportFormatters FormatterMap = FormatterMap{
	constants.CheckOutputFormatJSON: &JSONFormatter{},
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
	useFromSprigMap := []string{"upper", "toJson", "quote", "dict", "add", "now"}

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
	"steampipeversion": func() string { return version.SteampipeVersion.String() },
	"workingdir":       func() string { wd, _ := os.Getwd(); return wd },
	"timenow": func() string {
		return time.Now().Format(time.RFC3339)
	},
	"DurationInFloat": func(t time.Duration) float64 {
		return t.Seconds()
	},
}

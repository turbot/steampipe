package controldisplay

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/template"
	"time"

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
	constants.CheckOutputFormatCSV:   &CSVFormatter{},
	constants.CheckOutputFormatJSON:  &JSONFormatter{},
	constants.CheckOutputFormatText:  &TextFormatter{},
	constants.CheckOutputFormatBrief: &TextFormatter{},
}

var exportFormatters FormatterMap = FormatterMap{
	constants.CheckOutputFormatCSV:  &CSVFormatter{},
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

var formatterTemplateFuncMap template.FuncMap = template.FuncMap{
	"steampipeversion": func() string { return version.SteampipeVersion.String() },
	"workingdir":       func() string { wd, _ := os.Getwd(); return wd },
	"asstr":            func(i reflect.Value) string { return fmt.Sprintf("%v", i) },
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errors.New("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
	"summarystatusclass": func(status string, total int) string {
		switch strings.ToLower(status) {
		case "ok":
			if total > 0 {
				return "summary-total-ok highlight"
			}
			return "summary-total-ok"
		case "skip":
			if total > 0 {
				return "summary-total-skip highlight"
			}
			return "summary-total-skip"
		case "info":
			if total > 0 {
				return "summary-total-info highlight"
			}
			return "summary-total-info"
		case "alarm":
			if total > 0 {
				return "summary-total-alarm highlight"
			}
			return "summary-total-alarm"
		case "error":
			if total > 0 {
				return "summary-total-error highlight"
			}
			return "summary-total-error"
		}
		return ""
	},
	"ToUpper": func(text string) string {
		return strings.ToUpper(text)
	},
	"timenow": func() string {
		return time.Now().Format(time.RFC3339)
	},
	"GetDimensionRegion": func(row *controlexecute.ResultRow) string {
		if row.Dimensions[0].Key == "region" {
			return row.Dimensions[0].Value
		}
		return "ap-south-1"
	},
	"GetDimensionAccount": func(row *controlexecute.ResultRow) string {
		if row.Dimensions[0].Key == "account_id" {
			return row.Dimensions[0].Value
		}
		return row.Dimensions[1].Value
	},
	"DurationInFloat": func(t time.Duration) float64 {
		return t.Seconds()
	},
}

package controldisplay

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/control/controlexecute"
	"github.com/turbot/steampipe/version"
)

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
	constants.CheckOutputFormatNone:     &NullFormatter{},
	constants.CheckOutputFormatCSV:      &CSVFormatter{},
	constants.CheckOutputFormatJSON:     &JSONFormatter{},
	constants.CheckOutputFormatText:     &TextFormatter{},
	constants.CheckOutputFormatBrief:    &TextFormatter{},
	constants.CheckOutputFormatHTML:     &HTMLFormatter{},
	constants.CheckOutputFormatMarkdown: &MarkdownFormatter{},
}

var exportFormatters FormatterMap = FormatterMap{
	constants.CheckOutputFormatCSV:      &CSVFormatter{},
	constants.CheckOutputFormatJSON:     &JSONFormatter{},
	constants.CheckOutputFormatHTML:     &HTMLFormatter{},
	constants.CheckOutputFormatMarkdown: &MarkdownFormatter{},
	constants.CheckOutputFormatNUnit3:   &Nunit3Formatter{},
}

type CheckExportTarget struct {
	Format string
	File   string
	Error  error
}

func NewCheckExportTarget(format string, file string, err error) CheckExportTarget {
	return CheckExportTarget{
		Format: format,
		File:   file,
		Error:  err,
	}
}

type Formatter interface {
	Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error)
	FileExtension() string
}

func GetExportFormatter(exportFormat string) (Formatter, error) {
	formatter, found := exportFormatters[exportFormat]
	if !found {
		f, err := tryTemplateFormatter(exportFormat)
		if err != nil {
			return nil, fmt.Errorf("invalid export format '%s' - must be one of %s", exportFormat, exportFormatters.keys())
		}
		formatter = f
	}
	return formatter, nil
}

func GetOutputFormatter(outputFormat string) (Formatter, error) {
	formatter, found := outputFormatters[outputFormat]
	if !found {
		return nil, fmt.Errorf("invalid output format '%s' - must be one of %s", outputFormat, outputFormatters.keys())
	}
	return formatter, nil
}

func InferFormatFromExportFileName(filename string) (string, error) {
	extension := filepath.Ext(filename)
	switch extension {
	case ".csv":
		return constants.OutputFormatCSV, nil
	case ".json":
		return constants.OutputFormatJSON, nil
	case ".html", ".htm":
		return constants.CheckOutputFormatHTML, nil
	case ".md", ".markdown":
		return constants.CheckOutputFormatMarkdown, nil
	default:
		// could not infer format
		return "", fmt.Errorf("could not infer valid export format from filename '%s'", filename)
	}
}

func tryTemplateFormatter(exportFormat string) (*TemplateFormatter, error) {
	stat, err := os.Stat(exportFormat)
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("cannot parse directory")
	}
	template, err := template.ParseFiles(exportFormat)
	if err != nil {
		return nil, err
	}
	return &TemplateFormatter{template: template, outputExtension: "spex"}, nil
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
	"statusicon": func(status string) string {
		switch strings.ToLower(status) {
		case "ok":
			return "✅"
		case "skip":
			return "⇨"
		case "info":
			return "ℹ"
		case "alarm":
			return "❌"
		case "error":
			return "❗"
		}
		return ""
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
}

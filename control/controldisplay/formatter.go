package controldisplay

import (
	"context"
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
	constants.OutputFormatNone:     &NullFormatter{},
	constants.OutputFormatCSV:      &CSVFormatter{},
	constants.OutputFormatJSON:     &JSONFormatter{},
	constants.OutputFormatText:     &TextFormatter{},
	constants.OutputFormatBrief:    &TextFormatter{},
	constants.OutputFormatHTML:     &HTMLFormatter{},
	constants.OutputFormatMarkdown: &MarkdownFormatter{},
}

var exportFormatters FormatterMap = FormatterMap{
	constants.OutputFormatCSV:      &CSVFormatter{},
	constants.OutputFormatJSON:     &JSONFormatter{},
	constants.OutputFormatHTML:     &HTMLFormatter{},
	constants.OutputFormatMarkdown: &MarkdownFormatter{},
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
		return nil, fmt.Errorf("invalid export format '%s' - must be one of %s", exportFormat, exportFormatters.keys())
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
		return constants.OutputFormatHTML, nil
	case ".md", ".markdown":
		return constants.OutputFormatMarkdown, nil
	default:
		// could not infer format
		return "", fmt.Errorf("could not infer valid export format from filename '%s'", filename)
	}
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

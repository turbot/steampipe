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

const (
	OutputFormatNone     = "none"
	OutputFormatText     = "text"
	OutputFormatBrief    = "brief"
	OutputFormatCSV      = "csv"
	OutputFormatJSON     = "json"
	OutputFormatHTML     = "html"
	OutputFormatMarkdown = "md"
)

var outputFormatters FormatterMap = FormatterMap{
	OutputFormatNone:     &NullFormatter{},
	OutputFormatCSV:      &CSVFormatter{},
	OutputFormatJSON:     &JSONFormatter{},
	OutputFormatText:     &TextFormatter{},
	OutputFormatBrief:    &TextFormatter{},
	OutputFormatHTML:     &HTMLFormatter{},
	OutputFormatMarkdown: &MarkdownFormatter{},
}

var exportFormatters FormatterMap = FormatterMap{
	OutputFormatCSV:      &CSVFormatter{},
	OutputFormatJSON:     &JSONFormatter{},
	OutputFormatHTML:     &HTMLFormatter{},
	OutputFormatMarkdown: &MarkdownFormatter{},
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
		return OutputFormatCSV, nil
	case ".json":
		return OutputFormatJSON, nil
	case ".html", ".htm":
		return OutputFormatHTML, nil
	case ".md", ".markdown":
		return OutputFormatMarkdown, nil
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
	"steampipeversion": func() string { return version.String() },
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
}

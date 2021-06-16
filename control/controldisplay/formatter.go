package controldisplay

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/turbot/steampipe/control/execute"
)

const (
	OutputFormatNone  = "none"
	OutputFormatText  = "text"
	OutputFormatBrief = "brief"
	OutputFormatCSV   = "csv"
	OutputFormatJSON  = "json"
)

var outputFormatters map[string]Formatter = map[string]Formatter{
	OutputFormatNone:  &NullFormatter{},
	OutputFormatCSV:   &CSVFormatter{},
	OutputFormatJSON:  &JSONFormatter{},
	OutputFormatText:  &TextFormatter{},
	OutputFormatBrief: &TextFormatter{},
}

var exportFormatters map[string]Formatter = map[string]Formatter{
	OutputFormatCSV:  &CSVFormatter{},
	OutputFormatJSON: &JSONFormatter{},
}

type CheckExportFormat struct {
	Format string
	File   string
}

func NewCheckOutputFormat(format string, file string) CheckExportFormat {
	return CheckExportFormat{
		Format: format,
		File:   file,
	}
}

type Formatter interface {
	Format(ctx context.Context, tree *execute.ExecutionTree) (io.Reader, error)
}

func GetExportFormatter(exportFormat string) (Formatter, error) {
	formatter, found := exportFormatters[exportFormat]
	if !found {
		return nil, fmt.Errorf("invalid export format '%s' - must be one of json,csv", exportFormat)
	}
	return formatter, nil
}

func GetOutputFormatter(outputFormat string) (Formatter, error) {
	formatter, found := outputFormatters[outputFormat]
	if !found {
		return nil, fmt.Errorf("invalid output format '%s' - must be one of json,csv,text,brief,none", outputFormat)
	}
	return formatter, nil
}

func InferFormatFromExportFileName(filename string) string {
	extension := strings.TrimPrefix(filepath.Ext(filename), ".")
	switch extension {
	case "csv", "json":
		return extension
	default:
		// return blank, so that it fails when it looks
		// up the formatter when it's to format
		return ""
	}
}

// NullFormatter is to be used when no output is expected. It always returns a `io.Reader` which
// reads an empty string
type NullFormatter struct{}

func (j *NullFormatter) Format(ctx context.Context, tree *execute.ExecutionTree) (io.Reader, error) {
	return strings.NewReader(""), nil
}

package export

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type Target struct {
	exporter      Exporter
	filePath      string
	isNamedTarget bool
}

func (t *Target) Export(ctx context.Context, input ExportSourceData) (string, error) {
	if t.exporter == nil {
		return "", fmt.Errorf("exporter is nil")
	}
	err := t.exporter.Export(ctx, input, t.filePath)
	if err != nil {
		return "", err
	} else {
		exportPath := t.filePath
		if !filepath.IsAbs(exportPath) {
			pwd, _ := os.Getwd()
			exportPath = filepath.Join(pwd, exportPath)
		}
		return fmt.Sprintf("File exported to %s", exportPath), nil
	}
}

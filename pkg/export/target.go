package export

import (
	"context"
	"fmt"
	"os"
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
		pwd, _ := os.Getwd()
		return fmt.Sprintf("File exported to %s/%s", pwd, t.filePath), nil
	}
}

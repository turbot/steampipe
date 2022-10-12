package export

import "context"

type Target struct {
	exporter Exporter
	filePath string
}

func (t *Target) Export(ctx context.Context, input ExportSourceData) error {
	return t.exporter.Export(ctx, input, t.filePath)
}

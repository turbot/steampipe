package export

import "context"

// ExportSourceData is an interface implemented by all types which can be used as an input to an exporter
type ExportSourceData interface {
	IsExportSourceData()
}

type Exporter interface {
	Export(ctx context.Context, input ExportSourceData, destPath string) error
	FileExtension() string
	Name() string
	Alias() string
}

type ExporterBase struct{}

func (*ExporterBase) Alias() string {
	return ""
}

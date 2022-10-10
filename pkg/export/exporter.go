package export

// ExportSourceData is an interface implemented by all types which can be used as an input to an exporter
type ExportSourceData interface {
	IsExportSourceData()
}

type Exporter interface {
	Export(input ExportSourceData, destPath string) error
	FileExtension() string
	Name() string
}

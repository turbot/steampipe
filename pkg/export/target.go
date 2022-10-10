package export

type Target struct {
	exporter Exporter
	filePath string
}

func (t *Target) Export(input ExportSourceData) error {
	return t.exporter.Export(input, t.filePath)
}

package modconfig

type ReflectionDataItem struct {
	Column     string
	Value      string
	ColumnType string
}

// ReflectionDataPrivider :: a mod resource which can provide a map of reflection data
type ReflectionDataProvider interface {
	TableSchema() string
	GetData() []ReflectionDataItem
}

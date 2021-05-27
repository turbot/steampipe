package execute

type ResultColumns struct {
	AllColumns       []string
	GroupColumns     map[string]string
	ResultColumns    map[string]string
	DimensionColumns []string
	TagColumns       []string
}

func newResultColumns(e *ExecutionTree) *ResultColumns {
	groupColumns, groupColumnsKeyOrder := ResultGroup{}.CsvColumns()

	resultColumns, resultColumnsKeyOrder := ResultRow{}.CsvColumns()

	dimensionColumns := e.DimensionColorGenerator.getDimensionProperties()
	tagColumns := e.getAllTags()

	allColumns := []string{}
	allColumns = append(allColumns, groupColumnsKeyOrder...)
	allColumns = append(allColumns, resultColumnsKeyOrder...)
	allColumns = append(allColumns, dimensionColumns...)
	allColumns = append(allColumns, tagColumns...)

	return &ResultColumns{
		GroupColumns:     groupColumns,
		ResultColumns:    resultColumns,
		DimensionColumns: dimensionColumns,
		TagColumns:       tagColumns,
		AllColumns:       allColumns,
	}
}

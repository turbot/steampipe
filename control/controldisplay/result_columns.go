package controldisplay

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"sort"

	"github.com/turbot/steampipe/control/controlexecute"
)

type CsvColumnPair struct {
	fieldName  string
	columnName string
}

type ResultColumns struct {
	AllColumns       []string
	GroupColumns     []CsvColumnPair
	ResultColumns    []CsvColumnPair
	DimensionColumns []string
	TagColumns       []string
}

func newResultColumns(e *controlexecute.ExecutionTree) *ResultColumns {
	groupColumns := getCsvColumns(*e.Root)
	rowColumns := getCsvColumns(controlexecute.ResultRow{})

	dimensionColumns := e.DimensionColorGenerator.GetDimensionProperties()
	tagColumns := e.GetAllTags()

	sort.Strings(dimensionColumns)
	sort.Strings(tagColumns)
	sort.Slice(rowColumns, func(i, j int) bool {
		return rowColumns[i].fieldName < rowColumns[j].fieldName
	})

	allColumns := []string{}

	for _, gC := range groupColumns {
		allColumns = append(allColumns, gC.columnName)
	}
	for _, rC := range rowColumns {
		allColumns = append(allColumns, rC.columnName)
	}

	allColumns = append(allColumns, dimensionColumns...)
	allColumns = append(allColumns, tagColumns...)

	return &ResultColumns{
		GroupColumns:     groupColumns,
		ResultColumns:    rowColumns,
		DimensionColumns: dimensionColumns,
		TagColumns:       tagColumns,
		AllColumns:       allColumns,
	}
}

func getCsvColumns(item interface{}) []CsvColumnPair {
	columns := []CsvColumnPair{}

	t := reflect.TypeOf(item)
	val := reflect.ValueOf(item)
	for i := 0; i < val.NumField(); i++ {
		fieldName := val.Type().Field(i).Name
		field, _ := t.FieldByName(fieldName)
		tag, ok := field.Tag.Lookup("csv")
		if ok {
			columns = append(columns, CsvColumnPair{
				fieldName:  fieldName,
				columnName: tag,
			})
		}
	}

	if len(columns) == 0 {
		debug.PrintStack()
		panic(fmt.Errorf("getCsvColumns: given interface does not contain any CSV tags"))
	}

	return columns
}

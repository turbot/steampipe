package controldisplay

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"sort"
	"strings"

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
	sort.Slice(rowColumns[:], func(i, j int) bool {
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
			// split by comma
			csvAttrs := strings.Split(tag, ",")
			for _, csvAttr := range csvAttrs {
				// trim spaces from the sides
				csvAttr = strings.TrimSpace(csvAttr)

				// csvColumnName[:propertyNameOfValue]
				split := strings.SplitN(csvAttr, ":", 2)
				if len(split) > 1 {
					// is this a sub-property
					columns = append(columns, CsvColumnPair{
						fieldName:  fmt.Sprintf("%s.%s", fieldName, strings.TrimSpace(split[1])),
						columnName: strings.TrimSpace(split[0]),
					})
				} else {
					columns = append(columns, CsvColumnPair{
						fieldName:  fieldName,
						columnName: csvAttr,
					})
				}
			}
		}
	}

	if len(columns) == 0 {
		debug.PrintStack()
		panic(fmt.Errorf("getCsvColumns: given interface does not contain any CSV tags"))
	}

	return columns
}

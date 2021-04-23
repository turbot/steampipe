package db

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/turbot/go-kit/helpers"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
)

const TagColumn = "column"
const TagColumnType = "column_type"

func CreateReflectionTables(workspace *workspace.Workspace, client *Client) error {

	commonColumnSql := getColumnDefinitions(modconfig.CoreReflectionData{})

	// first reflect on each mod resource type to extract columns
	var createSql []string
	createSql = append(createSql, getTableCreateSql(modconfig.Control{}, "steampipe_controls", commonColumnSql))
	createSql = append(createSql, getTableCreateSql(modconfig.Query{}, "steampipe_queries", commonColumnSql))
	createSql = append(createSql, getTableCreateSql(modconfig.ControlGroup{}, "steampipe_control_groups", commonColumnSql))

	_, err := client.ExecuteSync(strings.Join(createSql, "\n"))
	if err != nil {
		return err
	}

	// now populate the tables
	var insertSql []string
	for _, control := range workspace.ControlMap {
		insertSql = append(insertSql, getTableInsertSql(control, "steampipe_controls"))
	}
	for _, query := range workspace.QueryMap {
		insertSql = append(insertSql, getTableInsertSql(query, "steampipe_queries"))
	}
	for _, controlGroup := range workspace.ControlGroupMap {
		insertSql = append(insertSql, getTableInsertSql(controlGroup, "steampipe_control_groups"))
	}
	if len(insertSql) > 0 {
		_, err = client.ExecuteSync(strings.Join(insertSql, ";\n"))
	}
	// TODO PROBLEM WITH LABELS - ARRAY POINTER
	return err
}

func getTableInsertSql(item modconfig.ReflectionDataItem, tableName string) string {

	// for each item there is core reflection data (i.e. reflection resource all items have)
	// and item specific reflection data
	// get the core reflection data values
	valuesCore, columnsCore := getReflectionColumnValues(item.GetCoreReflectionData())
	// get item specific reflection data values from the item
	valuesItem, columnsItem := getReflectionColumnValues(item)

	columns := append(columnsCore, columnsItem...)
	values := append(valuesCore, valuesItem...)
	insertSql := fmt.Sprintf(`insert into %s (%s) values(%s)`, tableName, strings.Join(columns, ","), strings.Join(values, ","))
	return insertSql
}

func getReflectionColumnValues(item interface{}) ([]string, []string) {
	var columns, values []string

	// dereference item in vcase it is a pointer
	item = helpers.DereferencePointer(item)

	val := reflect.ValueOf(helpers.DereferencePointer(item))
	t := reflect.TypeOf(item)

	for i := 0; i < val.NumField(); i++ {
		fieldName := val.Type().Field(i).Name
		field, _ := t.FieldByName(fieldName)

		column, ok := field.Tag.Lookup(TagColumn)
		if !ok {
			continue
		}
		_, ok = field.Tag.Lookup(TagColumnType)
		if !ok {
			continue
		}

		value, ok := helpers.GetFieldValueFromInterface(item, fieldName)

		// all fields will be pointers
		if !ok || helpers.DereferencePointer(value) == nil {
			continue
		}

		values = append(values, PgEscapeString(typeHelpers.ToString(value)))
		columns = append(columns, column)
	}
	return values, columns
}

func getTableCreateSql(s interface{}, tableName string, commonColumnSql []string) string {
	columnDefinitions := append(commonColumnSql, getColumnDefinitions(s)...)

	tableSql := fmt.Sprintf(`create temp table %s (
%s
);`, tableName, strings.Join(columnDefinitions, ",\n"))
	return tableSql
}

func getColumnDefinitions(item interface{}) []string {
	t := reflect.TypeOf(item)

	var columnDef []string
	val := reflect.ValueOf(item)
	for i := 0; i < val.NumField(); i++ {
		fieldName := val.Type().Field(i).Name
		field, _ := t.FieldByName(fieldName)

		column, ok := field.Tag.Lookup(TagColumn)
		if !ok {
			continue
		}
		columnType, ok := field.Tag.Lookup(TagColumnType)
		if !ok {
			continue
		}
		columnDef = append(columnDef, fmt.Sprintf("  %s  %s", column, columnType))

	}
	return columnDef
}

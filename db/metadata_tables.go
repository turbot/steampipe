package db

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/turbot/go-kit/helpers"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

const TagColumn = "column"
const TagColumnType = "column_type"

func UpdateMetadataTables(workspaceResources *modconfig.WorkspaceResourceMaps, client *Client) error {
	// get the create sql for each table type
	clearSql := getClearTablesSql()

	// now get sql to populate the tables
	insertSql := getTableInsertSql(workspaceResources)

	sql := []string{
		"begin;",
		clearSql,
		insertSql,
		"commit;",
	}
	_, err := client.ExecuteSync(strings.Join(sql, "\n"))

	return err
}

func CreateMetadataTables(workspaceResources *modconfig.WorkspaceResourceMaps, client *Client) error {
	// get the sql for columns which every table has
	commonColumnSql := getColumnDefinitions(modconfig.ResourceMetadata{})

	// get the create sql for each table type
	createSql := getCreateTablesSql(commonColumnSql)

	// now get sql to populate the tables
	insertSql := getTableInsertSql(workspaceResources)

	sql := []string{
		"begin;",
		createSql,
		insertSql,
		"commit;",
	}
	_, err := client.ExecuteSync(strings.Join(sql, "\n"))
	if err != nil {
		return fmt.Errorf("failed to create reflection tables: %v", err)
	}

	return nil
}

func getCreateTablesSql(commonColumnSql []string) string {
	var createSql []string
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.Control{}, "steampipe_controls", commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.Query{}, "steampipe_queries", commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.ControlGroup{}, "steampipe_control_groups", commonColumnSql))
	return strings.Join(createSql, "\n")
}

func getClearTablesSql() string {
	var clearSql = []string{
		"delete from steampipe_controls;",
		"delete from steampipe_queries;",
		"delete from steampipe_control_groups;",
	}
	return strings.Join(clearSql, "\n")
}

func getTableInsertSql(workspaceResources *modconfig.WorkspaceResourceMaps) string {
	var insertSql []string

	// the maps will have the same resource keyed by long and short name - avoid dupes
	resourcesAdded := make(map[string]bool)

	for _, control := range workspaceResources.ControlMap {
		if _, added := resourcesAdded[control.Name()]; !added {
			resourcesAdded[control.Name()] = true
			insertSql = append(insertSql, getTableInsertSqlForResource(control, "steampipe_controls"))
		}
	}
	for _, query := range workspaceResources.QueryMap {
		if _, added := resourcesAdded[query.Name()]; !added {
			resourcesAdded[query.Name()] = true
			insertSql = append(insertSql, getTableInsertSqlForResource(query, "steampipe_queries"))
		}
	}
	for _, controlGroup := range workspaceResources.ControlGroupMap {
		if _, added := resourcesAdded[controlGroup.Name()]; !added {
			resourcesAdded[controlGroup.Name()] = true
			insertSql = append(insertSql, getTableInsertSqlForResource(controlGroup, "steampipe_control_groups"))
		}
	}
	return strings.Join(insertSql, "\n")
}

func getTableCreateSqlForResource(s interface{}, tableName string, commonColumnSql []string) string {
	columnDefinitions := append(commonColumnSql, getColumnDefinitions(s)...)

	tableSql := fmt.Sprintf(`create temp table %s (
%s
);`, tableName, strings.Join(columnDefinitions, ",\n"))
	return tableSql
}

// get the sql column definitions for tagged properties of the item
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

func getTableInsertSqlForResource(item modconfig.ResourceWithMetadata, tableName string) string {

	// for each item there is core reflection data (i.e. reflection resource all items have)
	// and item specific reflection data
	// get the core reflection data values
	valuesCore, columnsCore := getColumnValues(item.GetMetadata())
	// get item specific reflection data values from the item
	valuesItem, columnsItem := getColumnValues(item)

	columns := append(columnsCore, columnsItem...)
	values := append(valuesCore, valuesItem...)
	insertSql := fmt.Sprintf(`insert into %s (%s) values(%s);`, tableName, strings.Join(columns, ","), strings.Join(values, ","))
	return insertSql
}

// use reflection to evaluate the column names and values from item - return as 2 separate arrays
func getColumnValues(item interface{}) ([]string, []string) {
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
		value = helpers.DereferencePointer(value)
		if !ok || value == nil {
			continue
		}

		values = append(values, pgValue(value))
		columns = append(columns, column)
	}
	return values, columns
}

// convert the value into a postgres format value which can used in an insert statement
func pgValue(item interface{}) string {
	rt := reflect.TypeOf(item)
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		s := reflect.ValueOf(item)

		var items []string
		for i := 0; i < s.Len(); i++ {
			element := s.Index(i).Interface()
			elementString := typeHelpers.ToString(element)
			items = append(items, elementString)
		}
		res := PgEscapeString(fmt.Sprintf(`{%s}`, strings.Join(items, ",")))
		return res
	default:
		return PgEscapeString(typeHelpers.ToString(item))
	}
}

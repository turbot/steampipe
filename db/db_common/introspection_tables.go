package db_common

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/turbot/go-kit/helpers"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
	"github.com/turbot/steampipe/utils"
	"github.com/zclconf/go-cty/cty"
)

// TagColumn is the tag used to specify the column name and type in the introspection tables
const TagColumn = "column"

func CreateIntrospectionTables(ctx context.Context, workspaceResources *modconfig.WorkspaceResourceMaps, session *DatabaseSession) error {
	utils.LogTime("db.CreateIntrospectionTables start")
	defer utils.LogTime("db.CreateIntrospectionTables end")

	// get the sql for columns which every table has
	commonColumnSql := getColumnDefinitions(modconfig.ResourceMetadata{})

	// get the create sql for each table type
	createSql := getCreateTablesSql(commonColumnSql)

	// now get sql to populate the tables
	insertSql := getTableInsertSql(workspaceResources)

	sql := []string{createSql, insertSql}

	_, err := session.Connection.ExecContext(ctx, strings.Join(sql, "\n"))
	if err != nil {
		return fmt.Errorf("failed to create introspection tables: %v", err)
	}

	// return context error - this enables calling code to respond to cancellation
	return ctx.Err()
}

func getCreateTablesSql(commonColumnSql []string) string {
	var createSql []string
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.Control{}, constants.IntrospectionTableControl, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.Query{}, constants.IntrospectionTableQuery, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.Benchmark{}, constants.IntrospectionTableBenchmark, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.Mod{}, constants.IntrospectionTableMod, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.Variable{}, constants.IntrospectionTableVariable, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.ReportContainer{}, constants.IntrospectionTableReport, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.ReportContainer{}, constants.IntrospectionTableContainer, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.ReportTable{}, constants.IntrospectionTableReportTable, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.ReportText{}, constants.IntrospectionTableReportText, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.ReportCounter{}, constants.IntrospectionTableReportCounter, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.ReportChart{}, constants.IntrospectionTableReportChart, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.ResourceReference{}, constants.IntrospectionTableReference, commonColumnSql))
	return strings.Join(createSql, "\n")
}

func getTableInsertSql(workspaceResources *modconfig.WorkspaceResourceMaps) string {
	var insertSql []string

	for _, control := range workspaceResources.Controls {
		insertSql = append(insertSql, getTableInsertSqlForResource(control, constants.IntrospectionTableControl))
	}
	for _, query := range workspaceResources.Queries {
		insertSql = append(insertSql, getTableInsertSqlForResource(query, constants.IntrospectionTableQuery))
	}
	for _, benchmark := range workspaceResources.Benchmarks {
		insertSql = append(insertSql, getTableInsertSqlForResource(benchmark, constants.IntrospectionTableBenchmark))
	}
	for _, mod := range workspaceResources.Mods {
		insertSql = append(insertSql, getTableInsertSqlForResource(mod, constants.IntrospectionTableMod))
	}
	for _, variable := range workspaceResources.Variables {
		insertSql = append(insertSql, getTableInsertSqlForResource(variable, constants.IntrospectionTableVariable))
	}
	for _, report := range workspaceResources.Reports {
		insertSql = append(insertSql, getTableInsertSqlForResource(report, constants.IntrospectionTableReport))
	}
	for _, container := range workspaceResources.ReportContainers {
		insertSql = append(insertSql, getTableInsertSqlForResource(container, constants.IntrospectionTableContainer))
	}
	for _, table := range workspaceResources.ReportTables {
		// TODO KAI delete when we are sure
		// skip anonymous panels
		//if table.IsAnonymous() {
		//	continue
		//}
		insertSql = append(insertSql, getTableInsertSqlForResource(table, constants.IntrospectionTableReportTable))
	}
	for _, text := range workspaceResources.ReportTexts {
		// skip anonymous panels
		//if text.IsAnonymous() {
		//	continue
		//}
		insertSql = append(insertSql, getTableInsertSqlForResource(text, constants.IntrospectionTableReportText))

	}
	for _, counter := range workspaceResources.ReportCounters {
		// skip anonymous panels
		//if counter.IsAnonymous() {
		//	continue
		//}
		insertSql = append(insertSql, getTableInsertSqlForResource(counter, constants.IntrospectionTableReportCounter))

	}
	for _, chart := range workspaceResources.ReportCharts {
		// skip anonymous panels
		//if chart.IsAnonymous() {
		//	continue
		//}
		insertSql = append(insertSql, getTableInsertSqlForResource(chart, constants.IntrospectionTableReportChart))

	}
	for _, reference := range workspaceResources.References {
		insertSql = append(insertSql, getTableInsertSqlForResource(reference, constants.IntrospectionTableReference))
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
		columnTag, ok := newColumnTag(field)
		if !ok {
			continue
		}
		columnDef = append(columnDef, fmt.Sprintf("  %s  %s", columnTag.Column, columnTag.ColumnType))
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
	if item == nil {
		return nil, nil
	}
	var columns, values []string

	// dereference item in vcase it is a pointer
	item = helpers.DereferencePointer(item)

	val := reflect.ValueOf(helpers.DereferencePointer(item))
	t := reflect.TypeOf(item)

	for i := 0; i < val.NumField(); i++ {
		fieldName := val.Type().Field(i).Name
		field, _ := t.FieldByName(fieldName)

		columnTag, ok := newColumnTag(field)
		if !ok {
			continue
		}

		value, ok := helpers.GetFieldValueFromInterface(item, fieldName)

		// all fields will be pointers
		value = helpers.DereferencePointer(value)
		if !ok || value == nil {
			continue
		}

		// formatIntrospectionTableValue escapes values, and for json columns, converts them into escaped JSON
		// ignore JSON conversion errors - trust that array values read from hcl will be convertable
		formattedValue, _ := formatIntrospectionTableValue(value, columnTag)
		values = append(values, formattedValue)
		columns = append(columns, columnTag.Column)
	}
	return values, columns
}

// convert the value into a postgres format value which can used in an insert statement
func formatIntrospectionTableValue(item interface{}, columnTag *ColumnTag) (string, error) {
	// special handling for cty.Type and cty.Value data
	switch t := item.(type) {
	// if the item is a cty value, we always represent it as json
	case cty.Value:
		if columnTag.ColumnType != "jsonb" {
			return "nil", fmt.Errorf("data for column %s is of type cty.Value so column type should be 'jsonb' but is actually %s", columnTag.Column, columnTag.ColumnType)
		}
		str, err := parse.CtyToJSON(t)
		if err != nil {
			return "", err
		}
		return PgEscapeString(str), nil
	case cty.Type:
		// if the item is a cty value, we always represent it as json
		if columnTag.ColumnType != "text" {
			return "nil", fmt.Errorf("data for column %s is of type cty.Type so column type should be 'text' but is actually %s", columnTag.Column, columnTag.ColumnType)
		}
		return PgEscapeString(t.FriendlyName()), nil
	}

	switch columnTag.ColumnType {
	case "jsonb":
		jsonBytes, err := json.Marshal(reflect.ValueOf(item).Interface())
		if err != nil {
			return "", err
		}

		res := PgEscapeString(string(jsonBytes))
		return res, nil
	case "integer", "numeric", "decimal", "boolean":
		return typeHelpers.ToString(item), nil
	default:
		// for string column, escape the data
		return PgEscapeString(typeHelpers.ToString(item)), nil
	}
}

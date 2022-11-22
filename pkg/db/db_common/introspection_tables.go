package db_common

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/type_conversion"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

// TagColumn is the tag used to specify the column name and type in the introspection tables
const TagColumn = "column"

func CreateIntrospectionTables(ctx context.Context, workspaceResources *modconfig.ResourceMaps, conn *pgx.Conn) error {
	// get the sql for columns which every table has
	commonColumnSql := getColumnDefinitions(modconfig.ResourceMetadata{})

	switch viper.GetString(constants.ArgIntrospection) {
	case constants.IntrospectionInfo:
		return populateAllIntrospectionTables(ctx, workspaceResources, conn, commonColumnSql)
	case constants.IntrospectionControl:
		return populateControlIntrospectionTables(ctx, workspaceResources, conn, commonColumnSql)
	default:
		// just create (but do not populate) the mod introspection table
		// - this is used to check if the session is initialised
		return createModIntrospectionTable(ctx, workspaceResources, conn, commonColumnSql)
	}
}

func createModIntrospectionTable(ctx context.Context, workspaceResources *modconfig.ResourceMaps, conn *pgx.Conn, commonColumnSql []string) error {
	utils.LogTime("db.CreateIntrospectionTables start")
	defer utils.LogTime("db.CreateIntrospectionTables end")

	// get the create sql for each table type
	modTableSql := getTableCreateSqlForResource(modconfig.Mod{}, constants.IntrospectionTableMod, commonColumnSql)

	_, err := conn.Exec(ctx, modTableSql)
	if err != nil {
		return fmt.Errorf("failed to create mod introspection table: %v", err)
	}

	// return context error - this enables calling code to respond to cancellation
	return ctx.Err()
}

func populateAllIntrospectionTables(ctx context.Context, workspaceResources *modconfig.ResourceMaps, conn *pgx.Conn, commonColumnSql []string) error {
	utils.LogTime("db.CreateIntrospectionTables start")
	defer utils.LogTime("db.CreateIntrospectionTables end")

	// get the create sql for each table type
	createSql := getCreateTablesSql(commonColumnSql)

	// now get sql to populate the tables
	insertSql := getTableInsertSql(workspaceResources)
	sql := []string{createSql, insertSql}

	_, err := conn.Exec(ctx, strings.Join(sql, "\n"))
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
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.Dashboard{}, constants.IntrospectionTableDashboard, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.DashboardContainer{}, constants.IntrospectionTableDashboardContainer, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.DashboardCard{}, constants.IntrospectionTableDashboardCard, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.DashboardChart{}, constants.IntrospectionTableDashboardChart, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.DashboardFlow{}, constants.IntrospectionTableDashboardFlow, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.DashboardGraph{}, constants.IntrospectionTableDashboardGraph, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.DashboardHierarchy{}, constants.IntrospectionTableDashboardHierarchy, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.DashboardImage{}, constants.IntrospectionTableDashboardImage, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.DashboardInput{}, constants.IntrospectionTableDashboardInput, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.DashboardTable{}, constants.IntrospectionTableDashboardTable, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.DashboardText{}, constants.IntrospectionTableDashboardText, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.ResourceReference{}, constants.IntrospectionTableReference, commonColumnSql))
	return strings.Join(createSql, "\n")
}

func getTableInsertSql(workspaceResources *modconfig.ResourceMaps) string {
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
		if !mod.IsDefaultMod() {
			insertSql = append(insertSql, getTableInsertSqlForResource(mod, constants.IntrospectionTableMod))
		}
	}
	for _, variable := range workspaceResources.Variables {
		insertSql = append(insertSql, getTableInsertSqlForResource(variable, constants.IntrospectionTableVariable))
	}
	for _, dashboard := range workspaceResources.Dashboards {
		insertSql = append(insertSql, getTableInsertSqlForResource(dashboard, constants.IntrospectionTableDashboard))
	}
	for _, container := range workspaceResources.DashboardContainers {
		insertSql = append(insertSql, getTableInsertSqlForResource(container, constants.IntrospectionTableDashboardContainer))
	}
	for _, card := range workspaceResources.DashboardCards {
		insertSql = append(insertSql, getTableInsertSqlForResource(card, constants.IntrospectionTableDashboardCard))
	}
	for _, chart := range workspaceResources.DashboardCharts {
		insertSql = append(insertSql, getTableInsertSqlForResource(chart, constants.IntrospectionTableDashboardChart))
	}
	for _, flow := range workspaceResources.DashboardFlows {
		insertSql = append(insertSql, getTableInsertSqlForResource(flow, constants.IntrospectionTableDashboardFlow))
	}
	for _, graph := range workspaceResources.DashboardGraphs {
		insertSql = append(insertSql, getTableInsertSqlForResource(graph, constants.IntrospectionTableDashboardGraph))
	}
	for _, hierarchy := range workspaceResources.DashboardHierarchies {
		insertSql = append(insertSql, getTableInsertSqlForResource(hierarchy, constants.IntrospectionTableDashboardHierarchy))
	}
	for _, image := range workspaceResources.DashboardImages {
		insertSql = append(insertSql, getTableInsertSqlForResource(image, constants.IntrospectionTableDashboardImage))
	}
	for _, dashboardInputs := range workspaceResources.DashboardInputs {
		for _, input := range dashboardInputs {
			insertSql = append(insertSql, getTableInsertSqlForResource(input, constants.IntrospectionTableDashboardInput))
		}
	}
	for _, input := range workspaceResources.GlobalDashboardInputs {
		insertSql = append(insertSql, getTableInsertSqlForResource(input, constants.IntrospectionTableDashboardInput))
	}
	for _, table := range workspaceResources.DashboardTables {
		insertSql = append(insertSql, getTableInsertSqlForResource(table, constants.IntrospectionTableDashboardTable))
	}
	for _, text := range workspaceResources.DashboardTexts {
		insertSql = append(insertSql, getTableInsertSqlForResource(text, constants.IntrospectionTableDashboardText))
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

func populateControlIntrospectionTables(ctx context.Context, workspaceResources *modconfig.ResourceMaps, conn *pgx.Conn, commonColumnSql []string) error {
	utils.LogTime("db.CreateIntrospectionTables start")
	defer utils.LogTime("db.CreateIntrospectionTables end")

	// get the create sql for control and benchmark tables
	createSql := getCreateControlTablesSql(commonColumnSql)
	// now get sql to populate the control and benchmark tables
	insertSql := getControlTableInsertSql(workspaceResources)
	sql := []string{createSql, insertSql}

	_, err := conn.Exec(ctx, strings.Join(sql, "\n"))
	if err != nil {
		return fmt.Errorf("failed to create introspection tables: %v", err)
	}

	// return context error - this enables calling code to respond to cancellation
	return ctx.Err()
}

func getCreateControlTablesSql(commonColumnSql []string) string {
	var createSql []string
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.Control{}, constants.IntrospectionTableControl, commonColumnSql))
	createSql = append(createSql, getTableCreateSqlForResource(modconfig.Benchmark{}, constants.IntrospectionTableBenchmark, commonColumnSql))
	return strings.Join(createSql, "\n")
}

func getControlTableInsertSql(workspaceResources *modconfig.ResourceMaps) string {
	var insertSql []string

	for _, control := range workspaceResources.Controls {
		insertSql = append(insertSql, getTableInsertSqlForResource(control, constants.IntrospectionTableControl))
	}
	for _, benchmark := range workspaceResources.Benchmarks {
		insertSql = append(insertSql, getTableInsertSqlForResource(benchmark, constants.IntrospectionTableBenchmark))
	}

	return strings.Join(insertSql, "\n")
}

// getColumnDefinitions returns the sql column definitions for tagged properties of the item
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
		str, err := type_conversion.CtyToJSON(t)
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

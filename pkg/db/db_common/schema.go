package db_common

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

type schemaRecord struct {
	TableSchema       string
	TableName         string
	ColumnName        string
	UdtName           string
	ColumnDefault     string
	IsNullable        string
	DataType          string
	ColumnDescription string
	TableDescription  string
}

func LoadForeignSchemaNames(ctx context.Context, conn *pgx.Conn) ([]string, error) {
	res, err := conn.Query(ctx, "SELECT DISTINCT foreign_table_schema FROM information_schema.foreign_tables WHERE foreign_server_name='steampipe'")
	if err != nil {
		return nil, err
	}

	var foreignSchemaNames []string
	var schema string
	for res.Next() {
		if err := res.Scan(&schema); err != nil {
			return nil, err
		}
		// ignore internal schema and legacy command schema
		if schema != constants.InternalSchema && schema != constants.LegacyCommandSchema {
			foreignSchemaNames = append(foreignSchemaNames, schema)
		}
	}
	sort.Strings(foreignSchemaNames)
	return foreignSchemaNames, nil
}

func LoadSchemaMetadata(ctx context.Context, conn *pgx.Conn, query string) (*SchemaMetadata, error) {
	var schemaRecords []schemaRecord
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schemaRecords, err = getSchemaRecordsFromRows(rows)
	if err != nil {
		return nil, err
	}

	// build schema metadata from query result
	return buildSchemaMetadata(schemaRecords)
}

func buildSchemaMetadata(records []schemaRecord) (_ *SchemaMetadata, err error) {
	utils.LogTime("db.buildSchemaMetadata start")
	defer func() {
		utils.LogTime("db.buildSchemaMetadata end")
	}()
	schemaMetadata := NewSchemaMetadata()

	utils.LogTime("db.buildSchemaMetadata.iteration start")
	for _, record := range records {
		if _, schemaFound := schemaMetadata.Schemas[record.TableSchema]; !schemaFound {
			schemaMetadata.Schemas[record.TableSchema] = map[string]TableSchema{}
		}

		if _, tblFound := schemaMetadata.Schemas[record.TableSchema][record.TableName]; !tblFound {
			schemaMetadata.Schemas[record.TableSchema][record.TableName] = TableSchema{
				Schema:      record.TableSchema,
				Name:        record.TableName,
				FullName:    fmt.Sprintf("%s.%s", record.TableSchema, record.TableName),
				Description: record.TableDescription,
				Columns:     map[string]ColumnSchema{},
			}
		}

		schemaMetadata.Schemas[record.TableSchema][record.TableName].Columns[record.ColumnName] = ColumnSchema{
			Name:        record.ColumnName,
			NotNull:     typeHelpers.StringToBool(record.IsNullable),
			Type:        record.DataType,
			Default:     record.ColumnDefault,
			Description: record.ColumnDescription,
		}

		if strings.HasPrefix(record.TableSchema, "pg_temp") {
			schemaMetadata.TemporarySchemaName = record.TableSchema
		}
	}
	utils.LogTime("db.buildSchemaMetadata.iteration end")

	return schemaMetadata, err
}

func getSchemaRecordsFromRows(rows pgx.Rows) ([]schemaRecord, error) {
	utils.LogTime("db.getSchemaRecordsFromRows start")
	defer utils.LogTime("db.getSchemaRecordsFromRows end")

	var records []schemaRecord

	// set this to the number of cols that are getting fetched
	numCols := 9

	rawResult := make([][]byte, numCols)
	dest := make([]interface{}, numCols) // A temporary interface{} slice
	for i := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	for rows.Next() {
		err := rows.Scan(dest...)
		if err != nil {
			return nil, err
		}

		t := schemaRecord{
			TableName:         string(rawResult[0]),
			ColumnName:        string(rawResult[1]),
			ColumnDefault:     string(rawResult[2]),
			IsNullable:        string(rawResult[3]),
			DataType:          string(rawResult[4]),
			UdtName:           string(rawResult[5]),
			TableSchema:       string(rawResult[6]),
			ColumnDescription: string(rawResult[7]),
			TableDescription:  string(rawResult[8]),
		}
		// for ltree data type, we need to use UdtName
		if t.DataType == "USER-DEFINED" {
			t.DataType = t.UdtName
		}
		records = append(records, t)
	}

	return records, nil
}

package db_common

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
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
		// ignore command schema
		if schema != constants.CommandSchema {
			foreignSchemaNames = append(foreignSchemaNames, schema)
		}
	}
	sort.Strings(foreignSchemaNames)
	return foreignSchemaNames, nil
}

func BuildSchemaMetadata(rows pgx.Rows) (_ *SchemaMetadata, err error) {
	utils.LogTime("db.buildSchemaMetadata start")
	defer func() {
		utils.LogTime("db.buildSchemaMetadata end")
		// ensure rows are closed
		rows.Close()
	}()

	records, err := getSchemaRecordsFromRows(rows)
	if err != nil {
		return nil, err
	}
	SchemaMetadata := NewSchemaMetadata()

	utils.LogTime("db.buildSchemaMetadata.iteration start")
	for _, record := range records {
		if _, schemaFound := SchemaMetadata.Schemas[record.TableSchema]; !schemaFound {
			SchemaMetadata.Schemas[record.TableSchema] = map[string]TableSchema{}
		}

		if _, tblFound := SchemaMetadata.Schemas[record.TableSchema][record.TableName]; !tblFound {
			SchemaMetadata.Schemas[record.TableSchema][record.TableName] = TableSchema{
				Schema:      record.TableSchema,
				Name:        record.TableName,
				FullName:    fmt.Sprintf("%s.%s", record.TableSchema, record.TableName),
				Description: record.TableDescription,
				Columns:     map[string]ColumnSchema{},
			}
		}

		SchemaMetadata.Schemas[record.TableSchema][record.TableName].Columns[record.ColumnName] = ColumnSchema{
			Name:        record.ColumnName,
			NotNull:     typeHelpers.StringToBool(record.IsNullable),
			Type:        record.DataType,
			Default:     record.ColumnDefault,
			Description: record.ColumnDescription,
		}

		if strings.HasPrefix(record.TableSchema, "pg_temp") {
			SchemaMetadata.TemporarySchemaName = record.TableSchema
		}
	}
	utils.LogTime("db.buildSchemaMetadata.iteration end")

	return SchemaMetadata, err
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

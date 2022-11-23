package db_common

import (
	"strings"

	"github.com/jackc/pgx/v5"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/schema"
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

func BuildSchemaMetadata(rows pgx.Rows) (_ *schema.Metadata, err error) {
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
	schemaMetadata := schema.NewMetadata()

	utils.LogTime("db.buildSchemaMetadata.iteration start")
	for _, record := range records {
		_, schemaFound := schemaMetadata.Schemas[record.TableSchema]
		if !schemaFound {
			schemaMetadata.Schemas[record.TableSchema] = map[string]schema.TableSchema{}
		}
		_, tblFound := schemaMetadata.Schemas[record.TableSchema][record.TableName]
		if !tblFound {
			schemaMetadata.Schemas[record.TableSchema][record.TableName] = schema.TableSchema{
				Name:        record.TableName,
				Schema:      record.TableSchema,
				Description: record.TableDescription,
				Columns:     map[string]schema.ColumnSchema{},
			}
		}

		schemaMetadata.Schemas[record.TableSchema][record.TableName].Columns[record.ColumnName] = schema.ColumnSchema{
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

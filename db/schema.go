package db

import (
	"database/sql"

	typeHelpers "github.com/turbot/go-kit/types"

	"github.com/turbot/steampipe/schema"
)

type schemaRecord struct {
	TableSchema       string
	TableName         string
	ColumnName        string
	ColumnDefault     string
	IsNullable        string
	DataType          string
	ColumnDescription string
	TableDescription  string
}

func buildSchemaMetadata(rows *sql.Rows) (*schema.Metadata, error) {
	records, err := getSchemaRecordsFromRows(rows)
	if err != nil {
		return nil, err
	}
	schemaMetadata := &schema.Metadata{
		Schemas: map[string]map[string]schema.TableSchema{},
	}

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
	}

	return schemaMetadata, err
}

func getSchemaRecordsFromRows(rows *sql.Rows) ([]schemaRecord, error) {
	records := []schemaRecord{}

	// set this to the number of cols that are getting fetched
	numCols := 8

	rawResult := make([][]byte, numCols)
	dest := make([]interface{}, numCols) // A temporary interface{} slice
	for i := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	for {
		if rows.Next() == false {
			break
		}

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
			TableSchema:       string(rawResult[5]),
			ColumnDescription: string(rawResult[6]),
			TableDescription:  string(rawResult[7]),
		}

		records = append(records, t)
	}

	return records, nil
}

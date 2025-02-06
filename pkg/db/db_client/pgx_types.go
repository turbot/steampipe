package db_client

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/turbot/pipe-fittings/v2/queryresult"
	"github.com/turbot/pipe-fittings/v2/utils"
)

// ColumnTypeDatabaseTypeName returns the database system type name. If the name is unknown the OID is returned.
func columnTypeDatabaseTypeName(field pgconn.FieldDescription, connection *pgx.Conn) (typeName string) {
	if dt, ok := connection.TypeMap().TypeForOID(field.DataTypeOID); ok {
		return strings.ToUpper(dt.Name)
	}

	return strconv.FormatInt(int64(field.DataTypeOID), 10)
}

func fieldDescriptionsToColumns(fieldDescriptions []pgconn.FieldDescription, connection *pgx.Conn) ([]*queryresult.ColumnDef, error) {
	cols := make([]*queryresult.ColumnDef, len(fieldDescriptions))

	for i, f := range fieldDescriptions {
		typeName := columnTypeDatabaseTypeName(f, connection)

		cols[i] = &queryresult.ColumnDef{
			Name:     string(f.Name),
			DataType: typeName,
		}
	}

	// Ensure column names are unique
	if err := ensureUniqueColumnName(cols); err != nil {
		return nil, err
	}

	return cols, nil
}

func ensureUniqueColumnName(cols []*queryresult.ColumnDef) error {
	// create a unique name generator
	nameGenerator := utils.NewUniqueNameGenerator()

	for colIdx, col := range cols {
		uniqueName, err := nameGenerator.GetUniqueName(col.Name, colIdx)
		if err != nil {
			return fmt.Errorf("error generating unique column name: %w", err)
		}
		// if the column name has changed, store the original name and update the column name to be the unique name
		if uniqueName != col.Name {
			// set the original name first, BEFORE mutating name
			col.OriginalName = col.Name
			col.Name = uniqueName
		}
	}
	return nil
}

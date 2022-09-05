package db_client

import (
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"
	"strconv"
	"strings"
)

// ColumnTypeDatabaseTypeName returns the database system type name. If the name is unknown the OID is returned.
func columnTypeDatabaseTypeName(field pgproto3.FieldDescription, connection *pgx.Conn) string {
	if dt, ok := connection.ConnInfo().DataTypeForOID(field.DataTypeOID); ok {
		return strings.ToUpper(dt.Name)
	}

	return strconv.FormatInt(int64(field.DataTypeOID), 10)
}

func fieldDescriptionsToColumns(fieldDescriptions []pgproto3.FieldDescription, connection *pgx.Conn) (cols, colTypes []string) {
	cols = make([]string, len(fieldDescriptions))
	colTypes = make([]string, len(fieldDescriptions))
	for i, f := range fieldDescriptions {
		cols[i] = string(f.Name)
		colTypes[i] = columnTypeDatabaseTypeName(f, connection)
	}
	return
}

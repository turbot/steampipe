package db_client

import (
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"
	"github.com/turbot/steampipe/pkg/query/queryresult"
	"reflect"
	"strconv"
	"strings"
)

// ColumnTypeDatabaseTypeName returns the database system type name. If the name is unknown the OID is returned.
func columnTypeDatabaseTypeName(field pgproto3.FieldDescription, connection *pgx.Conn) (typeName string, scanType reflect.Type) {
	if dt, ok := connection.ConnInfo().DataTypeForOID(field.DataTypeOID); ok {

		return strings.ToUpper(dt.Name), reflect.TypeOf(dt.Value)
	}

	return strconv.FormatInt(int64(field.DataTypeOID), 10), nil
}

func fieldDescriptionsToColumns(fieldDescriptions []pgproto3.FieldDescription, connection *pgx.Conn) []*queryresult.ColumnDef {
	cols := make([]*queryresult.ColumnDef, len(fieldDescriptions))

	for i, f := range fieldDescriptions {
		typeName, scanType := columnTypeDatabaseTypeName(f, connection)

		cols[i] = &queryresult.ColumnDef{
			Name:     string(f.Name),
			DataType: typeName,
			ScanType: scanType,
		}
	}
	return cols
}

package backend

import (
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/turbot/steampipe/pkg/query/queryresult"
	"github.com/turbot/steampipe/pkg/utils"
)

// SqliteRowReader is a RowReader implementation for the sqlite database/sql driver
type SqliteRowReader struct{}

func (r *SqliteRowReader) Read(columnValues []any, cols []*queryresult.ColumnDef) ([]any, error) {
	result := make([]any, len(columnValues))
	for i, columnValue := range columnValues {
		if columnValue != nil {
			result[i] = columnValue
			switch cols[i].DataType {
			case "_TEXT":
				if arr, ok := columnValue.([]interface{}); ok {
					elements := utils.Map(arr, func(e interface{}) string { return e.(string) })
					result[i] = strings.Join(elements, ",")
				}
			case "INET":
				if inet, ok := columnValue.(netip.Prefix); ok {
					result[i] = strings.TrimSuffix(inet.String(), "/32")
				}
			case "UUID":
				if bytes, ok := columnValue.([16]uint8); ok {
					if u, err := uuid.FromBytes(bytes[:]); err == nil {
						result[i] = u
					}
				}
			case "TIME":
				if t, ok := columnValue.(pgtype.Time); ok {
					result[i] = time.UnixMicro(t.Microseconds).UTC().Format("15:04:05")
				}
			case "INTERVAL":
				if interval, ok := columnValue.(pgtype.Interval); ok {
					var sb strings.Builder
					years := interval.Months / 12
					months := interval.Months % 12
					if years > 0 {
						sb.WriteString(fmt.Sprintf("%d %s ", years, utils.Pluralize("year", int(years))))
					}
					if months > 0 {
						sb.WriteString(fmt.Sprintf("%d %s ", months, utils.Pluralize("mon", int(months))))
					}
					if interval.Days > 0 {
						sb.WriteString(fmt.Sprintf("%d %s ", interval.Days, utils.Pluralize("day", int(interval.Days))))
					}
					if interval.Microseconds > 0 {
						d := time.Duration(interval.Microseconds) * time.Microsecond
						formatStr := time.Unix(0, 0).UTC().Add(d).Format("15:04:05")
						sb.WriteString(formatStr)
					}
					result[i] = sb.String()
				}

			case "NUMERIC":
				if numeric, ok := columnValue.(pgtype.Numeric); ok {
					if f, err := numeric.Float64Value(); err == nil {
						result[i] = f.Float64
					}
				}
			}
		}
	}
	return result, nil
}

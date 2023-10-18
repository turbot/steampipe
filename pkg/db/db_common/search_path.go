package db_common

import (
	"context"
	"database/sql"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
)

func EnsureInternalSchemaSuffix(searchPath []string) []string {
	// remove the InternalSchema
	searchPath = helpers.RemoveFromStringSlice(searchPath, constants.InternalSchema)
	// append the InternalSchema
	searchPath = append(searchPath, constants.InternalSchema)
	return searchPath
}

func AddSearchPathPrefix(searchPathPrefix []string, searchPath []string) []string {
	if len(searchPathPrefix) > 0 {
		prefixedSearchPath := searchPathPrefix
		for _, p := range searchPath {
			if !helpers.StringSliceContains(prefixedSearchPath, p) {
				prefixedSearchPath = append(prefixedSearchPath, p)
			}
		}
		searchPath = prefixedSearchPath
	}
	return searchPath
}

func BuildSearchPathResult(searchPathString string) ([]string, error) {
	// remove any leading/trailing braces
	searchPathString = strings.TrimPrefix(searchPathString, "{")
	searchPathString = strings.TrimSuffix(searchPathString, "}")
	// if this is called from GetSteampipeUserSearchPath the result will be prefixed by "search_path="
	searchPathString = strings.TrimPrefix(searchPathString, "search_path=")
	// split
	searchPath := strings.Split(searchPathString, ",")

	// unescape
	for idx, p := range searchPath {
		p = strings.Join(strings.Split(p, "\""), "")
		p = strings.TrimSpace(p)
		searchPath[idx] = p
	}
	return searchPath, nil
}

// TODO:: BINAEK :: we need to fix this
// this is going to be referred to from steampipe code in the future
// and
func GetUserSearchPath(ctx context.Context, conn *sql.Conn) ([]string, error) {
	return []string{}, nil
	// query := `SELECT array_to_string(rs.setconfig, ',')
	// FROM   pg_db_role_setting rs
	// LEFT   JOIN pg_roles      r ON r.oid = rs.setrole
	// LEFT   JOIN pg_database   d ON d.oid = rs.setdatabase
	// WHERE  r.rolname = 'steampipe'`

	// rows := conn.QueryRowContext(ctx, query)
	// var configStrings string
	// if err := rows.Scan(&configStrings); err != nil {
	// 	if errors.Is(err, sql.ErrNoRows) {
	// 		return []string{}, nil
	// 	}
	// 	return nil, err
	// }
	// if len(configStrings) > 0 {
	// 	return BuildSearchPathResult(configStrings)
	// }
	// // should not get here
	// return nil, nil
}

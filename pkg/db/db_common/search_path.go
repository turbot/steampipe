package db_common

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/v2/pkg/constants"
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
			if !slices.Contains(prefixedSearchPath, p) {
				prefixedSearchPath = append(prefixedSearchPath, p)
			}
		}
		searchPath = prefixedSearchPath
	}
	return searchPath
}

func BuildSearchPathResult(searchPathString string) ([]string, error) {
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

func GetUserSearchPath(ctx context.Context, conn *pgx.Conn) ([]string, error) {
	query := `SELECT rs.setconfig
	FROM   pg_db_role_setting rs
	LEFT   JOIN pg_roles      r ON r.oid = rs.setrole
	LEFT   JOIN pg_database   d ON d.oid = rs.setdatabase
	WHERE  r.rolname = 'steampipe'`

	rows := conn.QueryRow(ctx, query)
	var configStrings []string
	if err := rows.Scan(&configStrings); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []string{}, nil
		}
		return nil, err
	}
	if len(configStrings) > 0 {
		return BuildSearchPathResult(configStrings[0])
	}
	// should not get here
	return nil, nil
}

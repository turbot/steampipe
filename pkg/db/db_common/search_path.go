package db_common

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turbot/go-kit/helpers"
	"strings"
)

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

func GetUserSearchPath(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	query := `SELECT rs.setconfig
	FROM   pg_db_role_setting rs
	LEFT   JOIN pg_roles      r ON r.oid = rs.setrole
	LEFT   JOIN pg_database   d ON d.oid = rs.setdatabase
	WHERE  r.rolname = 'steampipe'`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var searchPathString string

		if err := rows.Scan(&searchPathString); err != nil {
			return nil, err
		}
		return BuildSearchPathResult("")
	}

	// should not get here
	return nil, nil
}

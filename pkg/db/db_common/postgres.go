package db_common

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// PgEscapeName escapes strings which will be usaed for Podsdtgres object identifiers
// (table names, column names, schema names)
func PgEscapeName(name string) string {
	// first escape all quotes by prefixing an addition quote
	name = strings.Replace(name, `"`, `""`, -1)
	// now wrap the whole string in quotes
	return fmt.Sprintf(`"%s"`, name)
}

// PgEscapeString escapes strings which are to be inserted
// use a custom escape tag to avoid chance of clash with the escaped text
// https://medium.com/@lnishada/postgres-dollar-quoting-6d23e4f186ec
func PgEscapeString(str string) string {
	return fmt.Sprintf(`$steampipe_escape$%s$steampipe_escape$`, str)
}

// PgEscapeSearchPath applies postgres escaping to search path and remove whitespace
func PgEscapeSearchPath(searchPath []string) []string {
	res := make([]string, len(searchPath))
	for idx, path := range searchPath {
		res[idx] = PgEscapeName(strings.TrimSpace(path))
	}
	return res
}

// PgIsInRecovery runs the pg_is_in_recovery over the connection provided.
// DOES NOT close the connection
func PgIsInRecovery(ctx context.Context, connection *pgx.Conn) (bool, error) {
	out := connection.QueryRow(ctx, "SELECT pg_is_in_recovery()")
	var IsInRecovery bool
	err := out.Scan(&IsInRecovery)
	if err != nil {
		return false, err
	}
	return IsInRecovery, nil
}

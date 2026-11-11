package db_common

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsAuthFailure reports whether postgres refused a login: SQLSTATE class 28, which covers
// "no pg_hba.conf entry", a rule that rejects the connection, and a wrong password.
func IsAuthFailure(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && strings.HasPrefix(pgErr.Code, "28")
}

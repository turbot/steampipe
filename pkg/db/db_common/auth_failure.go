package db_common

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsAuthFailure reports whether postgres refused a login: SQLSTATE class 28, which covers
// "no pg_hba.conf entry", a rule that rejects the connection, and a wrong password. A login for a
// role that does not exist has the same class, but no rule refused it.
func IsAuthFailure(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && strings.HasPrefix(pgErr.Code, "28") && !IsMissingRole(err)
}

// IsMissingRole reports whether postgres refused a login because the role does not exist.
func IsMissingRole(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "28000" &&
		strings.HasPrefix(pgErr.Message, "role ") && strings.HasSuffix(pgErr.Message, " does not exist")
}

package db_common

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"regexp"
)

func IsRelationNotFoundError(err error) bool {
	_, _, isRelationNotFound := GetMissingSchemaFromIsRelationNotFoundError(err)
	return isRelationNotFound
}

func GetMissingSchemaFromIsRelationNotFoundError(err error) (string, string, bool) {
	if err == nil {
		return "", "", false
	}
	var pgErr *pgconn.PgError
	ok := errors.As(err, &pgErr)
	if !ok || pgErr.Code != "42P01" {
		return "", "", false
	}

	r := regexp.MustCompile(`^relation "(.*)\.(.*)" does not exist$`)
	captureGroups := r.FindStringSubmatch(pgErr.Message)
	if len(captureGroups) == 3 {

		return captureGroups[1], captureGroups[2], true
	}

	// maybe there is no schema
	r = regexp.MustCompile(`^relation "(.*)" does not exist$`)
	captureGroups = r.FindStringSubmatch(pgErr.Message)
	if len(captureGroups) == 2 {
		return "", captureGroups[1], true
	}
	return "", "", true
}

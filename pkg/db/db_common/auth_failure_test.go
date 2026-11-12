package db_common

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsMissingRole(t *testing.T) {
	if !IsMissingRole(fmt.Errorf("connect: %w", &pgconn.PgError{Code: "28000", Message: `role "root" does not exist`})) {
		t.Error("a login for a role that does not exist")
	}
	if IsMissingRole(&pgconn.PgError{Code: "28000", Message: `no pg_hba.conf entry for host "[local]", user "root"`}) {
		t.Error("a refused login is not a missing role")
	}
	if IsMissingRole(errors.New(`role "root" does not exist`)) {
		t.Error("not a postgres error")
	}
}

func TestIsAuthFailure(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"no pg_hba.conf entry", &pgconn.PgError{Code: "28000"}, true},
		{"wrong password", &pgconn.PgError{Code: "28P01"}, true},
		{"the role does not exist", &pgconn.PgError{Code: "28000", Message: `role "root" does not exist`}, false},
		{"wrapped", fmt.Errorf("connect: %w", &pgconn.PgError{Code: "28000"}), true},
		{"the database is starting up", &pgconn.PgError{Code: "57P03"}, false},
		{"not a postgres error", errors.New("connection refused"), false},
		{"nil", nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsAuthFailure(c.err); got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

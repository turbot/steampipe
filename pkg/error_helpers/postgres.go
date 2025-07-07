package error_helpers

import (
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
)

func DecodePgError(err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return fmt.Errorf("%s", pgError.Message)
	}
	return err
}

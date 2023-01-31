package error_helpers

import (
	"errors"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/jackc/pgconn"
)

func DecodePgError(err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return fmt.Errorf(pgError.Message)
	}
	return err
}

func IsPreparedStatementDoesNotExistError(err error) bool {
	if err == nil {
		return false
	}
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return pgError.Code == "26000" && pgError.Routine == "FetchPreparedStatement"
	}

	return false
}

// PreparedStatementError is an error type to wrap PreparedStatement errors
type PreparedStatementError struct {
	queryName string

	underlying    error
	creationError error
	declRange     *hcl.Range
}

func NewPreparedStatementError(underlying error) *PreparedStatementError {
	return &PreparedStatementError{underlying: underlying}
}

func (e *PreparedStatementError) Error() string {
	// we may or may not have a query name - if not, just return the underlying
	if e.queryName == "" {
		return e.underlying.Error()
	}
	creationErrStr := ""
	if e.creationError != nil {
		var pgError *pgconn.PgError
		if errors.As(e.creationError, &pgError) {
			creationErrStr = fmt.Sprintf(": %s: %s", pgError.Message, e.declRange.String())
		} else {
			creationErrStr = fmt.Sprintf(": %s: %s", e.creationError.Error(), e.declRange.String())
		}
	}

	return fmt.Sprintf("error creating query '%s'%s", e.queryName, creationErrStr)
}

func (e *PreparedStatementError) Is(err error) bool {
	_, isPreparedStatementError := err.(*PreparedStatementError)
	return isPreparedStatementError
}

func (e *PreparedStatementError) Enrich(name string, err error, declRange *hcl.Range) *PreparedStatementError {
	e.queryName = name
	e.creationError = err
	e.declRange = declRange
	return e
}

// WrapPreparedStatementError modifies a context.Canceled error into a readable error that can
// be printed on the console
func WrapPreparedStatementError(err error) error {
	if IsPreparedStatementDoesNotExistError(err) {
		err = NewPreparedStatementError(err)
	}
	return err
}

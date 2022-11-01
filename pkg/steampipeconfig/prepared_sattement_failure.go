package steampipeconfig

import (
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type PreparedStatementFailure struct {
	Query modconfig.QueryProvider
	Error error
}

func (f *PreparedStatementFailure) String() string {
	if f.Query == nil {
		if pgErr, ok := f.Error.(*pgconn.PgError); ok {
			return fmt.Sprintf("failed to create all queries: %s", pgErr.Message)
		}
		return fmt.Sprintf("failed to create all queries: %s", f.Error)
	}
	return fmt.Sprintf("failed to create query '%s': %s (%s)", f.Query.Name(), f.Error, f.Query.GetDeclRange().String())
}

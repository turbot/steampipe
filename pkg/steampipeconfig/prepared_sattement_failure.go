package steampipeconfig

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type PreparedStatementFailure struct {
	Query modconfig.QueryProvider
	Error error
}

func (f *PreparedStatementFailure) String() string {
	if f.Query == nil {
		return fmt.Sprintf("failed to create all queries: %s", error_helpers.DecodePgError(f.Error).Error())
	}
	return fmt.Sprintf("failed to create query '%s': %s (%s)", f.Query.Name(), f.Error, f.Query.GetDeclRange().String())
}

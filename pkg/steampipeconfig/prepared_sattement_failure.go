package steampipeconfig

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type PreparedStatementFailure struct {
	Query *modconfig.Query
	Error error
}

func (f *PreparedStatementFailure) String() string {
	return fmt.Sprintf("failed to create query '%s': %s (%s)", f.Query.Name(), f.Error, f.Query.DeclRange.String())
}

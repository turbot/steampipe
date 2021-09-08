package db_common

import "github.com/turbot/steampipe/steampipeconfig/modconfig"

type QueryInitData struct {
	Queries                    []string
	PreparedStatementProviders *modconfig.WorkspaceResourceMaps
	Workspace                  WorkspaceResourceProvider
	Client                     Client
	Result                     *InitResult
}

func NewQueryInitData() *QueryInitData {
	return &QueryInitData{Result: &InitResult{}}
}

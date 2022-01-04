package query

import "github.com/turbot/steampipe/db/db_common"

type InitData struct {
	Queries   []string
	Workspace db_common.WorkspaceResourceProvider
	Client    db_common.Client
	Result    *db_common.InitResult
}

func NewInitData() *InitData {
	return &InitData{Result: &db_common.InitResult{}}
}

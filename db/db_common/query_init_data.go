package db_common

type QueryInitData struct {
	Queries   []string
	Workspace WorkspaceResourceProvider
	Client    Client
	Result    *InitResult
}

func NewQueryInitData() *QueryInitData {
	return &QueryInitData{Result: &InitResult{}}
}

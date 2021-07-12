package db

type QueryInitData struct {
	Queries   []string
	Workspace WorkspaceResourceProvider
	Client    *Client
	Result    *InitResult
}

func NewInitData() *QueryInitData {
	return &QueryInitData{Result: &InitResult{}}
}

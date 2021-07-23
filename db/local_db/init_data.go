package local_db

type QueryInitData struct {
	Queries   []string
	Workspace WorkspaceResourceProvider
	Client    *LocalClient
	Result    *InitResult
}

func NewInitData() *QueryInitData {
	return &QueryInitData{Result: &InitResult{}}
}

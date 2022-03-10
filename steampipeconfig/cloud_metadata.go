package steampipeconfig

type CloudMetadata struct {
	Actor            *ActorMetadata     `json:"actor,omitempty"`
	Identity         *IdentityMetadata  `json:"identity,omitempty"`
	Workspace        *WorkspaceMetadata `json:"workspace,omitempty"`
	ConnectionString string             `json:"-"`
}

func NewCloudMetadata() *CloudMetadata {
	return &CloudMetadata{
		Actor:     &ActorMetadata{},
		Identity:  &IdentityMetadata{},
		Workspace: &WorkspaceMetadata{},
	}
}

type ActorMetadata struct {
	Id     string `json:"id,omitempty"`
	Handle string `json:"handle,omitempty"`
}

type IdentityMetadata struct {
	Id     string `json:"id,omitempty"`
	Handle string `json:"handle,omitempty"`
	Type   string `json:"type,omitempty"`
}

type WorkspaceMetadata struct {
	Id     string `json:"id,omitempty"`
	Handle string `json:"handle,omitempty"`
}

package reportserver

type Server struct {
}

// Close closes the connection to the database and shuts down the backend
func (s *Server) Start() {
	StartAPI()
}

func (s *Server) HandleWorkspaceUpdate() {
	// TODO ...
}

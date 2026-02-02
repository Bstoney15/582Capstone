package server

import (
	"net/http"
	"backend/libraries/server/handlers"
	"backend/libraries/sessionmanager"
)

// server struct is used to inject dependencies into route handlers
type Server struct {
	Router         *http.ServeMux
	SessionManager *sessionmanager.SessionManager
}

// returnes a new server instance. all dependencies should be initialized here
func NewServer() *Server {
	return &Server{
		Router:         http.NewServeMux(),
		SessionManager: sessionmanager.NewSessionManager(),
	}
}

// start the servers and sets up any dependencies that are needed. 
func (s *Server) Start(address string) error {
	routes.AddRoutesToServer(s.Router)
	return http.ListenAndServe(address, s.Router)
}


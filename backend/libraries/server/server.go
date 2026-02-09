package server

import (
	routes "backend/libraries/server/handlers"
	"backend/libraries/sessionManager"
	"log"
	"net/http"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// server struct is used to inject dependencies into route handlers
type Server struct {
	Router         *http.ServeMux
	SessionManager *sessionManager.SessionManager
	DB             *gorm.DB
}

// returnes a new server instance. all dependencies should be initialized here
func NewServer() *Server {

	dsn := "root:secret@tcp(127.0.0.1:3306)/mysql?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	return &Server{
		Router:         http.NewServeMux(),
		SessionManager: sessionManager.NewSessionManager(),
		DB:             db,
	}
}

// start the servers and sets up any dependencies that are needed.
func (s *Server) Start(address string) error {
	routes.AddRoutesToServer(s.Router)
	return http.ListenAndServe(address, s.Router)
}

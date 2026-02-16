package routes

import (
	"backend/libraries/sessionManager"
	"net/http"

	"gorm.io/gorm"
)

type Handler struct {
	DB             *gorm.DB
	SessionManager *sessionManager.SessionManager
}

func NewHandler(db *gorm.DB, sm *sessionManager.SessionManager) *Handler {
	return &Handler{
		DB:             db,
		SessionManager: sm,
	}
}

/*
All api routes must be defined in this file. All routes must also be prefixed with api (/api/route)
try to group routes by functionality and authentication requirements

*/
func (h *Handler) RegisterRoutes(s *http.ServeMux) {

	// No Auth Routes
	s.HandleFunc("GET /api/health", h.HealthCheckHandler)

	s.HandleFunc("POST /api/user/login", h.LoginHandler)
	s.HandleFunc("POST /api/user/signup", h.SignupHandler)

	// Developer and above Routes

	
	// Admin Only Routes
}
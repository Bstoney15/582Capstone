package routes

import (
	"net/http"

	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		DB: db,
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
	s.HandleFunc("GET /api/user/auth", h.CheckAuthHandler)
	s.HandleFunc("GET /api/user/merchants", h.GetMerchantsHandler)
	s.HandleFunc("GET /api/user/info", h.GetUserInfo)

	// Developer and above Routes

	// Admin Only Routes
	s.HandleFunc("POST /api/merchant/create", h.CreateMerchantHandler)
	s.HandleFunc("POST /api/merchant/owner", h.CreateMerchantOwnerHandler)
	s.HandleFunc("POST /api/merchant/business-profile", h.CreateMerchantBusinessProfileHandler)
	s.HandleFunc("POST /api/merchant/address", h.CreateMerchantAddressHandler)
	s.HandleFunc("POST /api/merchant/add-user", h.AddUserHandler)
	s.HandleFunc("PATCH /api/merchant/edit-user-role", h.EditUserHandler)
	s.HandleFunc("GET /api/merchant/get-merchant-users", h.GetAllMerchantUsersHandler)
	s.HandleFunc("DELETE /api/merchant/remove-user", h.RemoveMerchantUserHandler)
}

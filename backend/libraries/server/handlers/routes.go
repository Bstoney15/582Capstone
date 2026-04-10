// Authors: Bem Stonestreet, Ryan Grimsley, Joe Hotze, Charley Findling
// Date Created: 02/02/26
// Description: file containing all routes for api backend
package routes

// Author: Benjamin Stonestreet
// Created: 2026-02-02

import (
	"net/http"

	"gorm.io/gorm"
)

// Handler encapsulates database dependencies for all API endpoints.
type Handler struct {
	DB *gorm.DB
}

// NewHandler initializes a new Handler with a database connection.
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		DB: db,
	}
}

/*
All api routes must be defined in this file. All routes must also be prefixed with api (/api/route)
try to group routes by functionality and authentication requirements
*/

// RegisterRoutes sets up all the HTTP route handlers on the provided ServeMux.
func (h *Handler) RegisterRoutes(s *http.ServeMux) {
	// Widget Static Routes
	s.HandleFunc("GET /widget/checkout.js", h.GetCheckoutWidget)
	s.HandleFunc("GET /api/widget/checkout.js", h.GetCheckoutWidget)

	// No Auth Routes
	s.HandleFunc("GET /api/health", h.HealthCheckHandler)
	s.HandleFunc("POST /api/invoices", h.CreateInvoiceHandler)
	s.HandleFunc("GET /api/invoices/{uuid}", h.GetInvoiceForCheckoutHandler)
	s.HandleFunc("GET /api/invoices/{uuid}/events", h.StreamInvoiceEventsHandler)
	s.HandleFunc("POST /api/verify", h.VerifyInvoicePaymentHandler)
	s.HandleFunc("OPTIONS /api/invoices", h.WidgetPreflightHandler)
	s.HandleFunc("OPTIONS /api/invoices/{uuid}", h.WidgetPreflightHandler)
	s.HandleFunc("OPTIONS /api/invoices/{uuid}/events", h.WidgetPreflightHandler)
	s.HandleFunc("OPTIONS /api/verify", h.WidgetPreflightHandler)

	s.HandleFunc("POST /api/user/login", h.LoginHandler)
	s.HandleFunc("POST /api/user/logout", h.LogoutHandler)
	s.HandleFunc("POST /api/user/signup", h.SignupHandler)
	s.HandleFunc("GET /api/user/auth", h.CheckAuthHandler)
	s.HandleFunc("GET /api/user/merchants", h.GetMerchantsHandler)
	s.HandleFunc("GET /api/user/info", h.GetUserInfo)
	s.HandleFunc("DELETE /api/user/merchants/{merchant_id}", h.LeaveMerchantHandler)

	// Developer and above Routes

	s.HandleFunc("GET /api/dashboard", h.GetDashboardHandler)
	s.HandleFunc("GET /api/dashboard/search", h.SearchInvoiceHandler)
	s.HandleFunc("GET /api/merchant/customers", h.GetMerchantCustomersHandler)

	// Admin Only Routes
	s.HandleFunc("POST /api/merchant/create", h.CreateMerchantHandler)
	s.HandleFunc("POST /api/merchant/add-user", h.AddUserHandler)
	s.HandleFunc("PATCH /api/merchant/edit-user-role", h.EditUserHandler)
	s.HandleFunc("GET /api/merchant/get-merchant-users", h.GetAllMerchantUsersHandler)
	s.HandleFunc("DELETE /api/merchant/remove-user", h.RemoveMerchantUserHandler)
	s.HandleFunc("GET /api/merchant/get-wallet", h.GetMerchantWalletHandler)
	s.HandleFunc("PATCH /api/merchant/set-wallet", h.SetMerchantWalletHandler)

	// Merchant API Key Routes (all merchant roles)
	s.HandleFunc("GET /api/merchant/api_key", h.GetMerchantAPIKeysHandler)
	s.HandleFunc("POST /api/merchant/api_key", h.CreateMerchantAPIKeyHandler)
	s.HandleFunc("DELETE /api/merchant/api_key/{api_key}", h.DeleteMerchantAPIKeyHandler)
}

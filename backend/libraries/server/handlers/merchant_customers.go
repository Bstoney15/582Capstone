// Authors: Connor Williamson
// Created: 04/08/26
// Description: handler functions for merchant customers
package routes

import (
	"backend/libraries/sessionManager"
	"backend/models"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"gorm.io/gorm"
)

func (h *Handler) GetMerchantCustomersHandler(w http.ResponseWriter, r *http.Request) {
	sessionToken, err := sessionManager.GetSessionToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionData, _, active := sessionManager.CheckSession(sessionToken)
	if !active {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	merchantID := r.URL.Query().Get("merchant_id")
	if merchantID == "" {
		http.Error(w, "merchant_id query parameter required", http.StatusBadRequest)
		return
	}

	// Verify the requesting user has any role for this merchant
	var role models.Role
	if err := h.DB.Where(
		"role_merchant_id = ? AND role_user_id = ? AND role_name IN ?",
		merchantID,
		sessionData.UserID,
		[]string{models.RoleDeveloper, models.RoleAdmin, models.RoleOwner},
	).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	query := h.DB.Model(&models.Customer{}).Where("customer_merchant_id = ?", merchantID)

	search := r.URL.Query().Get("search")
	if search != "" {
		like := fmt.Sprintf("%%%s%%", search)
		query = query.Where(
			"customer_first_name LIKE ? OR customer_last_name LIKE ? OR customer_email LIKE ?",
			like, like, like,
		)
	}

	var customers []models.Customer
	if err := query.Find(&customers).Error; err != nil {
		http.Error(w, "Failed to fetch customers", http.StatusInternalServerError)
		return
	}

	type CustomerResponse struct {
		CustomerID string `json:"customer_id"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		Email      string `json:"email"`
	}

	responses := make([]CustomerResponse, 0, len(customers))
	for _, c := range customers {
		responses = append(responses, CustomerResponse{
			CustomerID: c.CustomerID,
			FirstName:  c.CustomerFirstName,
			LastName:   c.CustomerLastName,
			Email:      c.CustomerEmail,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}

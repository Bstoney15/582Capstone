package routes

import (
	"backend/libraries/sessionManager"
	"backend/models"
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

type MerchantResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func (h *Handler) GetMerchantsHandler(w http.ResponseWriter, r *http.Request) {
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

	var roles []models.Role
	if err := h.DB.Preload("Merchant").Where("role_user_id = ?", sessionData.UserID).Find(&roles).Error; err != nil {
		http.Error(w, "Failed to fetch merchants", http.StatusInternalServerError)
		return
	}

	merchantResponses := make([]MerchantResponse, 0, len(roles))
	for _, role := range roles {
		if role.Merchant.MerchantID != "" {
			merchantResponses = append(merchantResponses, MerchantResponse{
				ID:   role.Merchant.MerchantID,
				Name: role.Merchant.MerchantName,
				Role: role.RoleName,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(merchantResponses)
}

func (h *Handler) AddUserHandler(w http.ResponseWriter, r *http.Request) {
	// auth copied from GetMerchantsHandler func
	sessionToken, err := sessionManager.GetSessionToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// may need to replace first _ with sessionData
	_, _, active := sessionManager.CheckSession(sessionToken)
	if !active {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// TODO add logic for adding user...

}

func (h *Handler) EditUserHandler(w http.ResponseWriter, r *http.Request) {
	// auth copied from GetMerchantsHandler func
	sessionToken, err := sessionManager.GetSessionToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// may need to replace first _ with sessionData
	_, _, active := sessionManager.CheckSession(sessionToken)
	if !active {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// TODO add logic for editing user...
}

func (h *Handler) GetAllMerchantUsers(w http.ResponseWriter, r *http.Request) {
	// auth copied from GetMerchantsHandler func
	sessionToken, err := sessionManager.GetSessionToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// may need to replace first _ with sessionData
	sessionData, _, active := sessionManager.CheckSession(sessionToken)
	if !active {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// TODO add logic for getting all users associated with a given merchant
	merchantID := r.URL.Query().Get("merchant_id")
	if merchantID == "" {
		http.Error(w, "merchant_id query parameter needed", http.StatusBadRequest)
		return
	}
	// check if user is actually a part of the merchants users for the merchant whos users are being requested
	// also check if user has role with priveledge to access all users list, currently i said only admin and owner can see users
	var users_role models.Role
	if err := h.DB.Where("role_merchant_id = ? AND role_user_id = ? and role_name IN ?", merchantID, sessionData.UserID, []string{models.RoleAdmin, models.RoleOwner}).First(&users_role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Failed", http.StatusNotFound)
			return
		}
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	// get the users of the merchant
	var roles []models.Role
	if err := h.DB.Preload("User").Where("role_merchant_id = ?", merchantID).Find(&roles).Error; err != nil {
		http.Error(w, "Search for users failed", http.StatusInternalServerError)
		return
	}
	// format users correctly and write to response
	type MerchantUserResponse struct {
		UserID     string `json:"user_id"`
		Username   string `json:"username"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		UserStatus string `json:"user_status"`
		Role       string `json:"role"`
	}

	responses := make([]MerchantUserResponse, 0, len(roles))
	for _, role := range roles {
		responses = append(responses, MerchantUserResponse{
			UserID:     role.User.UserID,
			Username:   role.User.UserUsername,
			FirstName:  role.User.UserFirstName,
			LastName:   role.User.UserLastName,
			UserStatus: role.User.UserStatus,
			Role:       role.RoleName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}

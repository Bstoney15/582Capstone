package routes

import (
	"backend/libraries/sessionManager"
	"backend/models"
	"encoding/json"
	"net/http"
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

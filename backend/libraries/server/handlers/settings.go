// Author: Charley Findling
// Created: 3/29/2026
// Description: Model for user settings, with a handler function to remove a user from a merchant at their request.


package routes

import (
	"backend/libraries/sessionManager"
	"backend/models"
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

// LeaveMerchantHandler removes the calling user's role entry for the given merchant,
// effectively unlinking the merchant from their account.
func (h *Handler) LeaveMerchantHandler(w http.ResponseWriter, r *http.Request) {
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

	merchantID := r.PathValue("merchant_id")
	if merchantID == "" {
		http.Error(w, "merchant_id is required", http.StatusBadRequest)
		return
	}

	// Find the role entry connecting this user to the merchant
	var role models.Role
	if err := h.DB.Where("role_merchant_id = ? AND role_user_id = ?", merchantID, sessionData.UserID).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "You are not associated with this merchant", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to look up role", http.StatusInternalServerError)
		return
	}

	// Delete the role entry
	if err := h.DB.Delete(&role).Error; err != nil {
		http.Error(w, "Failed to remove merchant association", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Merchant removed from your account",
	})
}
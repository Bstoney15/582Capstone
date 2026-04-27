// Author: Ryan Grimsley
// Date Created: 02/23/26
// Description: File containing handler to get user info, this runs when the route GET /api/user/info is hit
package routes

import (
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

func (h *Handler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	user, err := h.getAuthenticatedUser(r)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userProfileResponse(*user))

}

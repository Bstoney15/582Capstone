package routes

import (
	"backend/libraries/sessionManager"
	"backend/models"
	"encoding/json"
	"net/http"
)

// struct representing what to return of user's info
type UserInfoResponse struct {
	UserID     string `json:"user_id"`
	Username   string `json:"user_username"`
	FirstName  string `json:"user_first_name"`
	LastName   string `json:"user_last_name"`
	UserStatus string `json:"user_status"`
}

func (h *Handler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	//copied auth from GetMerchantsHandler
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

	// do query for users where user id matches user id of current session
	var users []models.User
	if err := h.DB.Where("user_id = ?", sessionData.UserID).Find(&users).Error; err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// handle if empty
	if len(users) == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	// convert user info to struct
	user := users[0]
	resp := UserInfoResponse{
		UserID:     user.UserID,
		Username:   user.UserUsername,
		FirstName:  user.UserFirstName,
		LastName:   user.UserLastName,
		UserStatus: user.UserStatus,
	}

	// add to respose writer
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

}

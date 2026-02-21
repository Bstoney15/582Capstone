package routes

import (
	"backend/libraries/sessionManager"
	"encoding/json"
	"net/http"
)

type AuthResponse struct {
	Authenticated bool   `json:"authenticated"`
	UserID        string `json:"userId,omitempty"`
}

func (h *Handler) CheckAuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sessionToken, err := sessionManager.GetSessionToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthResponse{Authenticated: false})
		return
	}

	sessionData, rotatedToken, active := sessionManager.CheckSession(sessionToken)
	if !active {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthResponse{Authenticated: false})
		return
	}

	if rotatedToken != "" {
		sessionManager.SetSessionCookie(w, rotatedToken)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{
		Authenticated: true,
		UserID:        sessionData.UserID,
	})
}

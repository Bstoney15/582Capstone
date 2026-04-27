// login.go – handler that authenticates a user with username and password and creates a session.
package routes

// Author: Benjamin Stonestreet
// Created: 2026-02-15

import (
	"backend/libraries/sessionManager"
	"backend/models"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// LoginRequest is the JSON payload expected by LoginHandler.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginHandler authenticates the user with the provided credentials, creates a session,
// and sets the session cookie on the response.
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Look up the user by username; a missing record is treated as an auth failure.
	var user models.User
	if err := h.DB.Where("user_username = ?", req.Username).First(&user).Error; err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Verify the plaintext password against the stored bcrypt hash.
	if err := bcrypt.CompareHashAndPassword([]byte(user.UserPasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Create a new session and set the cookie on the response.
	sessionID := sessionManager.CreateSession(user.UserID)
	sessionManager.SetSessionCookie(w, sessionID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user": map[string]string{
			"username":  user.UserUsername,
			"firstName": user.UserFirstName,
			"lastName":  user.UserLastName,
		},
	})
}

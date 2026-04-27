// signup.go – handler that registers a new user account and creates an initial session.
package routes

// Author: Benjamin Stonestreet
// Created: 2026-02-15

import (
	"backend/libraries/sessionManager"
	"backend/models"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// SignupRequest is the JSON payload expected by SignupHandler.
type SignupRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

// SignupHandler creates a new user account, bcrypt-hashes the password, starts a session,
// and returns the new user's basic info.
func (h *Handler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if username already exists.
	var existingUser models.User
	if err := h.DB.Where("user_username = ?", req.Username).First(&existingUser).Error; err == nil {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user := models.User{
		UserID:           uuid.New().String(),
		UserUsername:     req.Username,
		UserFirstName:    req.FirstName,
		UserLastName:     req.LastName,
		UserPasswordHash: string(hashedPassword),
		UserStatus:       "active",
	}

	if err := h.DB.Create(&user).Error; err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Create a session immediately so the user is logged in after signup.
	sessionID := sessionManager.CreateSession(user.UserID)
	sessionManager.SetSessionCookie(w, sessionID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Signup successful",
		"user": map[string]string{
			"username":  user.UserUsername,
			"firstName": user.UserFirstName,
			"lastName":  user.UserLastName,
		},
	})
}

// Author: Connor-Williamson
// Date Created: 2026-04-26
// Description: Handles user profile retrieval and updates with authentication validation, including username uniqueness checks and password verification.
package routes

import (
	"backend/libraries/sessionManager"
	"backend/models"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserProfileResponse struct {
	UserID     string `json:"user_id"`
	Username   string `json:"user_username"`
	FirstName  string `json:"user_first_name"`
	LastName   string `json:"user_last_name"`
	UserStatus string `json:"user_status"`
}

type UpdateUserProfileRequest struct {
	Username        *string `json:"user_username"`
	FirstName       *string `json:"user_first_name"`
	LastName        *string `json:"user_last_name"`
	CurrentPassword string  `json:"current_password"`
}

func userProfileResponse(user models.User) UserProfileResponse {
	return UserProfileResponse{
		UserID:     user.UserID,
		Username:   user.UserUsername,
		FirstName:  user.UserFirstName,
		LastName:   user.UserLastName,
		UserStatus: user.UserStatus,
	}
}

func (h *Handler) getAuthenticatedUser(r *http.Request) (*models.User, error) {
	sessionToken, err := sessionManager.GetSessionToken(r)
	if err != nil {
		return nil, err
	}

	sessionData, _, active := sessionManager.CheckSession(sessionToken)
	if !active {
		return nil, errors.New("inactive session")
	}

	var user models.User
	if err := h.DB.Where("user_id = ?", sessionData.UserID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *Handler) GetUserProfileHandler(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) UpdateUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.getAuthenticatedUser(r)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateUserProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updates := map[string]interface{}{}

	if req.Username != nil {
		username := strings.TrimSpace(*req.Username)
		if username == "" {
			http.Error(w, "user_username cannot be empty", http.StatusBadRequest)
			return
		}

		if username != user.UserUsername {
			var existingUser models.User
			err := h.DB.Where("user_username = ?", username).First(&existingUser).Error
			if err == nil {
				http.Error(w, "Username already taken", http.StatusConflict)
				return
			}
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "Failed to validate username", http.StatusInternalServerError)
				return
			}
		}

		updates["user_username"] = username
		user.UserUsername = username
	}

	if req.FirstName != nil {
		firstName := strings.TrimSpace(*req.FirstName)
		if firstName == "" {
			http.Error(w, "user_first_name cannot be empty", http.StatusBadRequest)
			return
		}

		updates["user_first_name"] = firstName
		user.UserFirstName = firstName
	}

	if req.LastName != nil {
		lastName := strings.TrimSpace(*req.LastName)
		if lastName == "" {
			http.Error(w, "user_last_name cannot be empty", http.StatusBadRequest)
			return
		}

		updates["user_last_name"] = lastName
		user.UserLastName = lastName
	}

	if len(updates) == 0 {
		http.Error(w, "No profile fields provided", http.StatusBadRequest)
		return
	}

	currentPassword := strings.TrimSpace(req.CurrentPassword)
	if currentPassword == "" {
		http.Error(w, "current_password is required", http.StatusBadRequest)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.UserPasswordHash), []byte(req.CurrentPassword)); err != nil {
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}

	if err := h.DB.Model(&models.User{}).Where("user_id = ?", user.UserID).Updates(updates).Error; err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userProfileResponse(*user))
}
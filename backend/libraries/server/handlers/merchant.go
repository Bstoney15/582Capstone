package routes

import (
	"backend/libraries/sessionManager"
	"backend/models"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
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

// defines if these roles are allowed to be set
var allowedRolesToSet = map[string]bool{
	models.RoleAdmin: true, models.RoleDeveloper: true, models.RoleOwner: false,
}

func (h *Handler) AddUserHandler(w http.ResponseWriter, r *http.Request) {
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

	type AddUserRequest struct {
		MerchantID   string `json:"merchant_id"`
		UserUsername string `json:"user_username"`
		Role         string `json:"role"`
	}

	var requestBody AddUserRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}

	if requestBody.MerchantID == "" || requestBody.UserUsername == "" || requestBody.Role == "" {
		http.Error(w, "merchant_id, user_username, and role are required", http.StatusBadRequest)
		return
	}

	if !allowedRolesToSet[requestBody.Role] {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	// Require caller to be Admin or Owner for this merchant
	var callerRole models.Role
	if err := h.DB.Where("role_merchant_id = ? AND role_user_id = ? AND role_name IN ?",
		requestBody.MerchantID, sessionData.UserID, []string{models.RoleAdmin, models.RoleOwner}).First(&callerRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	// Verify merchant exists
	var merchant models.Merchant
	if err := h.DB.Where("merchant_id = ?", requestBody.MerchantID).First(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Merchant not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify merchant", http.StatusInternalServerError)
		return
	}

	var user models.User
	if err := h.DB.Where("user_username = ?", requestBody.UserUsername).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "No user found with that username", http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to lookup user", http.StatusInternalServerError)
		return
	}

	// Avoid duplicate: user already has a role for this merchant
	var existing models.Role
	if err := h.DB.Where("role_merchant_id = ? AND role_user_id = ?", requestBody.MerchantID, user.UserID).First(&existing).Error; err == nil {
		http.Error(w, "User already has a role for this merchant", http.StatusConflict)
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "Failed to check existing role", http.StatusInternalServerError)
		return
	}

	role := &models.Role{
		RoleID:         uuid.New().String(),
		RoleMerchantID: requestBody.MerchantID,
		RoleUserID:     user.UserID,
		RoleName:       requestBody.Role,
	}
	if err := h.DB.Create(role).Error; err != nil {
		http.Error(w, "Failed to add user to merchant", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"role_id": role.RoleID,
		"message": "User added to merchant",
	})
}

// defines if these roles are allowed to edit other users roles
var rolesAllowedToEdit = map[string]bool{
	models.RoleAdmin: true, models.RoleDeveloper: false, models.RoleOwner: true,
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

	type EditUserRequest struct {
		MerchantID string `json:"merchant_id"`
		UserID     string `json:"user_id"`
		Role       string `json:"role"`
		EditorId   string `json:"editor_id"`
	}

	// try to parse request body
	var requestBody EditUserRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}
	// make sure request has all required params
	if requestBody.MerchantID == "" || requestBody.UserID == "" || requestBody.Role == "" {
		http.Error(w, "merchant_id, user_id, and role are required", http.StatusBadRequest)
		return
	}
	// get user who is trying to make an edit's role
	var editorRole models.Role
	if err := h.DB.Where("role_user_id = ? AND role_merchant_id = ?", requestBody.EditorId, requestBody.MerchantID).First(&editorRole).Error; err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	// check if that role is allowed to edit
	if !rolesAllowedToEdit[editorRole.RoleName] {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	// Verify merchant exists
	var merchant models.Merchant
	if err := h.DB.Where("merchant_id = ?", requestBody.MerchantID).First(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Merchant not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify merchant", http.StatusInternalServerError)
		return
	}
	// get user were trying to edit
	var user models.User
	if err := h.DB.Where("user_id = ?", requestBody.UserID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "No user found with that username", http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to lookup user", http.StatusInternalServerError)
		return
	}
	// get role associated with user
	var userRoleToEdit models.Role
	if err := h.DB.Where("role_merchant_id = ? AND role_user_id = ?", requestBody.MerchantID, user.UserID).First(&userRoleToEdit).Error; err != nil {
		http.Error(w, "failed to find users role", http.StatusNotFound)
		return
	}
	// edit the users role
	userRoleToEdit.RoleName = requestBody.Role
	h.DB.Save(&userRoleToEdit)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User edited",
	})

}

func (h *Handler) GetAllMerchantUsersHandler(w http.ResponseWriter, r *http.Request) {
	// auth copied from GetMerchantsHandler func
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

func (h *Handler) RemoveMerchantUserHandler(w http.ResponseWriter, r *http.Request) {
	sessionToken, err := sessionManager.GetSessionToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, _, active := sessionManager.CheckSession(sessionToken)
	if !active {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// TODO actually implement removing

}

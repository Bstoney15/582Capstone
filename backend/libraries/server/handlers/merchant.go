// merchant.go – handlers for merchant creation, listing, user management, and wallet operations.
package routes

// Author: Ryan Grimsley, Ben Stonestreet
// Created: 2026-02-20

import (
	"backend/libraries/sessionManager"
	"backend/models"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// MerchantResponse is the summary shape returned by GetMerchantsHandler for each merchant the caller belongs to.
type MerchantResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// CreateMerchantHandler creates a new merchant along with its business profile, address, and beneficial owner.
// All records are written inside a single DB transaction so the create is atomic.
func (h *Handler) CreateMerchantHandler(w http.ResponseWriter, r *http.Request) {
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

	type CreateMerchantRequest struct {
		MerchantName string `json:"merchant_name"`

		MerchantBusinessProfileDBAName            string `json:"merchant_business_profile_dba_name"`
		MerchantBusinessProfileRegistrationNumber string `json:"merchant_business_profile_registration_number"`
		MerchantBusinessProfileTaxID              string `json:"merchant_business_profile_tax_id"`
		MerchantBusinessProfileWebsiteURL         string `json:"merchant_business_profile_website_url"`
		MerchantBusinessProfileIncoporationDate   string `json:"merchant_business_profile_incoporation_date"`
		MerchantBusinessProfileLegalStructure     string `json:"merchant_business_profile_legal_structure"`
		MerchantBusinessProfileMCCCode            string `json:"merchant_business_profile_mcc_code"`
		MerchantBusinessProfilePhoneNumber        string `json:"merchant_business_profile_phone_number"`
		MerchantBusinessProfileEmail              string `json:"merchant_business_profile_email"`

		MerchantAddressLine1      string `json:"merchant_address_line_1"`
		MerchantAddressLine2      string `json:"merchant_address_line_2"`
		MerchantAddressCity       string `json:"merchant_address_city"`
		MerchantAddressState      string `json:"merchant_address_state"`
		MerchantAddressPostalCode int    `json:"merchant_address_postal_code"`

		MerchantOwnerFirstName   string `json:"merchant_owner_first_name"`
		MerchantOwnerLastName    string `json:"merchant_owner_last_name"`
		MerchantOwnerPhoneNumber string `json:"merchant_owner_phone_number"`
		MerchantOwnerEmail       string `json:"merchant_owner_email"`
		MerchantOwnerDOB         string `json:"merchant_owner_dob"`
		MerchantOwnerStake       string `json:"merchant_owner_stake"`
	}

	var requestBody CreateMerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}

	// All fields are required; reject the request if any are missing.
	if requestBody.MerchantName == "" ||
		requestBody.MerchantBusinessProfileDBAName == "" ||
		requestBody.MerchantBusinessProfileRegistrationNumber == "" ||
		requestBody.MerchantBusinessProfileTaxID == "" ||
		requestBody.MerchantBusinessProfileWebsiteURL == "" ||
		requestBody.MerchantBusinessProfileIncoporationDate == "" ||
		requestBody.MerchantBusinessProfileLegalStructure == "" ||
		requestBody.MerchantBusinessProfileMCCCode == "" ||
		requestBody.MerchantBusinessProfilePhoneNumber == "" ||
		requestBody.MerchantBusinessProfileEmail == "" ||
		requestBody.MerchantAddressLine1 == "" ||
		requestBody.MerchantAddressLine2 == "" ||
		requestBody.MerchantAddressCity == "" ||
		requestBody.MerchantAddressState == "" ||
		requestBody.MerchantAddressPostalCode == 0 ||
		requestBody.MerchantOwnerFirstName == "" ||
		requestBody.MerchantOwnerLastName == "" ||
		requestBody.MerchantOwnerPhoneNumber == "" ||
		requestBody.MerchantOwnerEmail == "" ||
		requestBody.MerchantOwnerDOB == "" ||
		requestBody.MerchantOwnerStake == "" {
		http.Error(w, "All merchant, business profile, address, and owner fields are required", http.StatusBadRequest)
		return
	}

	parsedOwnerDOB, err := time.Parse("2006-01-02", requestBody.MerchantOwnerDOB)
	if err != nil {
		http.Error(w, "merchant_owner_dob must be in YYYY-MM-DD format", http.StatusBadRequest)
		return
	}

	parsedOwnerStake, err := decimal.NewFromString(requestBody.MerchantOwnerStake)
	if err != nil {
		http.Error(w, "merchant_owner_stake must be a valid decimal string", http.StatusBadRequest)
		return
	}

	var caller models.User
	if err := h.DB.Where("user_id = ?", sessionData.UserID).First(&caller).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Failed to verify caller", http.StatusInternalServerError)
		return
	}

	merchantID := uuid.New().String()
	roleID := uuid.New().String()
	profileID := uuid.New().String()
	addressID := uuid.New().String()
	ownerID := uuid.New().String()

	// Write all five records atomically; a failure in any step rolls the entire transaction back.
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		merchant := models.Merchant{
			MerchantID:   merchantID,
			MerchantName: requestBody.MerchantName,
		}
		if err := tx.Create(&merchant).Error; err != nil {
			return err
		}

		ownerRole := models.Role{
			RoleID:         roleID,
			RoleMerchantID: merchantID,
			RoleUserID:     sessionData.UserID,
			RoleName:       models.RoleOwner,
		}
		if err := tx.Create(&ownerRole).Error; err != nil {
			return err
		}

		profile := models.MerchantBusinessProfile{
			MerchantBusinessProfileID:                 profileID,
			MerchantBusinessProfileMerchantID:         merchantID,
			MerchantBusinessProfileDBAName:            requestBody.MerchantBusinessProfileDBAName,
			MerchantBusinessProfileRegistrationNumber: requestBody.MerchantBusinessProfileRegistrationNumber,
			MerchantBusinessProfileTaxID:              requestBody.MerchantBusinessProfileTaxID,
			MerchantBusinessProfileWebsiteURL:         requestBody.MerchantBusinessProfileWebsiteURL,
			MerchantBusinessProfileIncoporationDate:   requestBody.MerchantBusinessProfileIncoporationDate,
			MerchantBusinessProfileLegalStructure:     requestBody.MerchantBusinessProfileLegalStructure,
			MerchantBusinessProfileMCCCode:            requestBody.MerchantBusinessProfileMCCCode,
			MerchantBusinessProfilePhoneNumber:        requestBody.MerchantBusinessProfilePhoneNumber,
			MerchantBusinessProfileEmail:              requestBody.MerchantBusinessProfileEmail,
		}
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}

		address := models.MerchantAddress{
			MerchantAddressID:         addressID,
			MerchantAddressMerchantID: merchantID,
			MerchantAddressLine1:      requestBody.MerchantAddressLine1,
			MerchantAddressLine2:      requestBody.MerchantAddressLine2,
			MerchantAddressCity:       requestBody.MerchantAddressCity,
			MerchantAddressState:      requestBody.MerchantAddressState,
			MerchantAddressPostalCode: requestBody.MerchantAddressPostalCode,
			MerchantAddressVerified:   false,
		}
		if err := tx.Create(&address).Error; err != nil {
			return err
		}

		owner := models.MerchantOwner{
			MerchantOwnerID:          ownerID,
			MerchantOwnerMerchantID:  merchantID,
			MerchantOwnerFirstName:   requestBody.MerchantOwnerFirstName,
			MerchantOwnerLastName:    requestBody.MerchantOwnerLastName,
			MerchantOwnerPhoneNumber: requestBody.MerchantOwnerPhoneNumber,
			MerchantOwnerEmail:       requestBody.MerchantOwnerEmail,
			MerchantOwnerDOB:         parsedOwnerDOB,
			MerchantOwnerStake:       parsedOwnerStake,
		}
		if err := tx.Create(&owner).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		http.Error(w, "Failed to create merchant", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"merchant_id":                  merchantID,
		"role_id":                      roleID,
		"merchant_business_profile_id": profileID,
		"merchant_address_id":          addressID,
		"merchant_owner_id":            ownerID,
		"role":                         models.RoleOwner,
		"message":                      "Merchant created with profile, address, and owner",
	})
}

// GetMerchantsHandler returns all merchants the authenticated caller belongs to, along with their role.
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

// allowedRolesToSet defines which roles can be assigned by AddUserHandler.
// Owner cannot be self-assigned to prevent privilege escalation.
var allowedRolesToSet = map[string]bool{
	models.RoleAdmin: true, models.RoleDeveloper: true, models.RoleOwner: false,
}

// AddUserHandler adds a user to a merchant with the given role. The caller must be an Admin or Owner.
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

	// Require caller to be Admin or Owner for this merchant.
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

	// Verify merchant exists.
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

	// Avoid duplicate: user already has a role for this merchant.
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

// rolesAllowedToEdit defines which roles are permitted to modify other users' roles.
var rolesAllowedToEdit = map[string]bool{
	models.RoleAdmin: true, models.RoleDeveloper: false, models.RoleOwner: true,
}

// EditUserHandler updates the role of an existing user within a merchant.
// The editor must hold a role that is in rolesAllowedToEdit.
func (h *Handler) EditUserHandler(w http.ResponseWriter, r *http.Request) {
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

	type EditUserRequest struct {
		MerchantID string `json:"merchant_id"`
		UserID     string `json:"user_id"`
		Role       string `json:"role"`
		EditorID   string `json:"editor_id"`
	}

	var requestBody EditUserRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}

	if requestBody.MerchantID == "" || requestBody.UserID == "" || requestBody.Role == "" || requestBody.EditorID == "" {
		http.Error(w, "merchant_id, user_id, and role are required", http.StatusBadRequest)
		return
	}

	// Look up the editor's role to determine if they have edit permission.
	var editorRole models.Role
	if err := h.DB.Where("role_user_id = ? AND role_merchant_id = ?", requestBody.EditorID, requestBody.MerchantID).First(&editorRole).Error; err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if !rolesAllowedToEdit[editorRole.RoleName] {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	// Verify merchant exists.
	var merchant models.Merchant
	if err := h.DB.Where("merchant_id = ?", requestBody.MerchantID).First(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Merchant not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify merchant", http.StatusInternalServerError)
		return
	}

	// Look up the user being edited.
	var user models.User
	if err := h.DB.Where("user_id = ?", requestBody.UserID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "No user found with that username", http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to lookup user", http.StatusInternalServerError)
		return
	}

	// Fetch the user's existing role record and update it.
	var userRoleToEdit models.Role
	if err := h.DB.Where("role_merchant_id = ? AND role_user_id = ?", requestBody.MerchantID, user.UserID).First(&userRoleToEdit).Error; err != nil {
		http.Error(w, "failed to find users role", http.StatusNotFound)
		return
	}

	userRoleToEdit.RoleName = requestBody.Role
	h.DB.Save(&userRoleToEdit)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User edited",
	})

}

// GetAllMerchantUsersHandler returns all users associated with the specified merchant.
// The caller must be an Admin or Owner of the merchant.
func (h *Handler) GetAllMerchantUsersHandler(w http.ResponseWriter, r *http.Request) {
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

	// Confirm the caller has a privileged role before exposing the user list.
	var users_role models.Role
	if err := h.DB.Where("role_merchant_id = ? AND role_user_id = ? and role_name IN ?", merchantID, sessionData.UserID, []string{models.RoleAdmin, models.RoleOwner}).First(&users_role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Failed", http.StatusNotFound)
			return
		}
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	var roles []models.Role
	if err := h.DB.Preload("User").Where("role_merchant_id = ?", merchantID).Find(&roles).Error; err != nil {
		http.Error(w, "Search for users failed", http.StatusInternalServerError)
		return
	}

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

// RemoveMerchantUserHandler removes a user's role from a merchant, revoking their access.
// The editor must hold a role that is in rolesAllowedToEdit.
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

	type RemoveUserRequest struct {
		MerchantID string `json:"merchant_id"`
		UserID     string `json:"user_id"`
		EditorID   string `json:"editor_id"`
	}

	var requestBody RemoveUserRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}

	if requestBody.MerchantID == "" || requestBody.UserID == "" || requestBody.EditorID == "" {
		http.Error(w, "merchant_id, user_id, and editor_id are required", http.StatusBadRequest)
		return
	}

	// Verify the editor has sufficient privilege.
	var removerRole models.Role
	if err := h.DB.Where("role_user_id = ? AND role_merchant_id = ?", requestBody.EditorID, requestBody.MerchantID).First(&removerRole).Error; err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if !rolesAllowedToEdit[removerRole.RoleName] {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	// Verify merchant exists.
	var merchant models.Merchant
	if err := h.DB.Where("merchant_id = ?", requestBody.MerchantID).First(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Merchant not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify merchant", http.StatusInternalServerError)
		return
	}

	// Fetch the role record to delete.
	var userRole models.Role
	if err := h.DB.Where("role_user_id = ?", requestBody.UserID).First(&userRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "No user found", http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to lookup user", http.StatusInternalServerError)
		return
	}

	h.DB.Delete(&userRole)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User removed",
	})
}

// GetMerchantWalletHandler returns the XRPL wallet address for the specified merchant.
// The caller must be an Admin or Owner.
func (h *Handler) GetMerchantWalletHandler(w http.ResponseWriter, r *http.Request) {
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

	// Confirm the caller has a privileged role for this merchant.
	var users_role models.Role
	if err := h.DB.Where("role_merchant_id = ? AND role_user_id = ? and role_name IN ?", merchantID, sessionData.UserID, []string{models.RoleAdmin, models.RoleOwner}).First(&users_role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Failed", http.StatusNotFound)
			return
		}
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	var wallet models.MerchantCryptoWallet
	if err := h.DB.Where("merchant_crypto_wallet_merchant_id = ?", merchantID).First(&wallet).Error; err != nil {
		http.Error(w, "Error retrieving wallet", http.StatusInternalServerError)
		return
	}

	type WalletResponse struct {
		WalletAddress string `json:"wallet_address"`
	}

	resp := WalletResponse{
		WalletAddress: wallet.MerchantCryptoWalletAddress,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// SetMerchantWalletHandler updates the XRPL wallet address for the specified merchant.
// The caller must be an Admin or Owner.
func (h *Handler) SetMerchantWalletHandler(w http.ResponseWriter, r *http.Request) {
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

	// Confirm the caller has a privileged role for this merchant.
	var users_role models.Role
	if err := h.DB.Where("role_merchant_id = ? AND role_user_id = ? and role_name IN ?", merchantID, sessionData.UserID, []string{models.RoleAdmin, models.RoleOwner}).First(&users_role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Failed", http.StatusNotFound)
			return
		}
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	type SetWalletRequest struct {
		WalletAddress string `json:"wallet_address"`
	}

	var requestBody SetWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}

	if requestBody.WalletAddress == "" {
		http.Error(w, "wallet_address required", http.StatusBadRequest)
		return
	}

	// Fetch the existing wallet record and update the address in place.
	//! this assumes the entry for the merchants wallet exists, not sure if that will be the real way its implemented
	// for ex, if new merchant account by default has no crytpo wallet entry
	var MerchantCryptoWallet models.MerchantCryptoWallet
	if err := h.DB.Where("merchant_crypto_wallet_merchant_id = ?", merchantID).First(&MerchantCryptoWallet).Error; err != nil {
		http.Error(w, "failed to find merchant crypto wallet entry", http.StatusNotFound)
		return
	}

	MerchantCryptoWallet.MerchantCryptoWalletAddress = requestBody.WalletAddress
	h.DB.Save(&MerchantCryptoWallet)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Wallet address set",
	})

}

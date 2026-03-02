package routes

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

type MerchantResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

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
	}

	var requestBody CreateMerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}

	if requestBody.MerchantName == "" {
		http.Error(w, "merchant_name is required", http.StatusBadRequest)
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

		return nil
	})

	if err != nil {
		http.Error(w, "Failed to create merchant", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"merchant_id": merchantID,
		"role_id":     roleID,
		"role":        models.RoleOwner,
		"message":     "Merchant created and owner role assigned",
	})
}

func (h *Handler) requireMerchantAccess(merchantID string, userID string) error {
	if merchantID == "" || userID == "" {
		return errors.New("missing merchant_id or user_id")
	}

	var merchant models.Merchant
	if err := h.DB.Where("merchant_id = ?", merchantID).First(&merchant).Error; err != nil {
		return err
	}

	var callerRole models.Role
	if err := h.DB.Where("role_merchant_id = ? AND role_user_id = ?", merchantID, userID).First(&callerRole).Error; err != nil {
		return err
	}

	return nil
}

func (h *Handler) CreateMerchantOwnerHandler(w http.ResponseWriter, r *http.Request) {
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

	type CreateMerchantOwnerRequest struct {
		MerchantID  string `json:"merchant_id"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		PhoneNumber string `json:"phone_number"`
		Email       string `json:"email"`
		DOB         string `json:"dob"`
		Stake       string `json:"stake"`
	}

	var requestBody CreateMerchantOwnerRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}

	if requestBody.MerchantID == "" || requestBody.FirstName == "" || requestBody.LastName == "" || requestBody.PhoneNumber == "" || requestBody.Email == "" || requestBody.DOB == "" || requestBody.Stake == "" {
		http.Error(w, "merchant_id, first_name, last_name, phone_number, email, dob, and stake are required", http.StatusBadRequest)
		return
	}

	if err := h.requireMerchantAccess(requestBody.MerchantID, sessionData.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to verify merchant and permissions", http.StatusInternalServerError)
		return
	}

	parsedDOB, err := time.Parse("2006-01-02", requestBody.DOB)
	if err != nil {
		http.Error(w, "dob must be in YYYY-MM-DD format", http.StatusBadRequest)
		return
	}

	parsedStake, err := decimal.NewFromString(requestBody.Stake)
	if err != nil {
		http.Error(w, "stake must be a valid decimal string", http.StatusBadRequest)
		return
	}

	var owner models.MerchantOwner
	err = h.DB.Where("merchant_owner_merchant_id = ?", requestBody.MerchantID).First(&owner).Error
	if err == nil {
		owner.MerchantOwnerFirstName = requestBody.FirstName
		owner.MerchantOwnerLastName = requestBody.LastName
		owner.MerchantOwnerPhoneNumber = requestBody.PhoneNumber
		owner.MerchantOwnerEmail = requestBody.Email
		owner.MerchantOwnerDOB = parsedDOB
		owner.MerchantOwnerStake = parsedStake
		if err := h.DB.Save(&owner).Error; err != nil {
			http.Error(w, "Failed to update merchant owner", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"merchant_owner_id": owner.MerchantOwnerID,
			"message":           "Merchant owner updated",
		})
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "Failed to check existing merchant owner", http.StatusInternalServerError)
		return
	}

	owner = models.MerchantOwner{
		MerchantOwnerID:          uuid.New().String(),
		MerchantOwnerMerchantID:  requestBody.MerchantID,
		MerchantOwnerFirstName:   requestBody.FirstName,
		MerchantOwnerLastName:    requestBody.LastName,
		MerchantOwnerPhoneNumber: requestBody.PhoneNumber,
		MerchantOwnerEmail:       requestBody.Email,
		MerchantOwnerDOB:         parsedDOB,
		MerchantOwnerStake:       parsedStake,
	}

	if err := h.DB.Create(&owner).Error; err != nil {
		http.Error(w, "Failed to create merchant owner", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"merchant_owner_id": owner.MerchantOwnerID,
		"message":           "Merchant owner created",
	})
}

func (h *Handler) CreateMerchantBusinessProfileHandler(w http.ResponseWriter, r *http.Request) {
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

	type CreateMerchantBusinessProfileRequest struct {
		MerchantID         string `json:"merchant_id"`
		DBAName            string `json:"dba_name"`
		RegistrationNumber string `json:"registration_number"`
		TaxID              string `json:"tax_id"`
		WebsiteURL         string `json:"website_url"`
		IncoporationDate   string `json:"incoporation_date"`
		LegalStructure     string `json:"legal_structure"`
		MCCCode            string `json:"mcc_code"`
		PhoneNumber        string `json:"phone_number"`
		Email              string `json:"email"`
	}

	var requestBody CreateMerchantBusinessProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}

	if requestBody.MerchantID == "" || requestBody.DBAName == "" || requestBody.RegistrationNumber == "" || requestBody.TaxID == "" || requestBody.WebsiteURL == "" || requestBody.IncoporationDate == "" || requestBody.LegalStructure == "" || requestBody.MCCCode == "" || requestBody.PhoneNumber == "" || requestBody.Email == "" {
		http.Error(w, "merchant_id, dba_name, registration_number, tax_id, website_url, incoporation_date, legal_structure, mcc_code, phone_number, and email are required", http.StatusBadRequest)
		return
	}

	if err := h.requireMerchantAccess(requestBody.MerchantID, sessionData.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to verify merchant and permissions", http.StatusInternalServerError)
		return
	}

	var profile models.MerchantBusinessProfile
	err = h.DB.Where("merchant_business_profile_merchant_id = ?", requestBody.MerchantID).First(&profile).Error
	if err == nil {
		profile.MerchantBusinessProfileDBAName = requestBody.DBAName
		profile.MerchantBusinessProfileRegistrationNumber = requestBody.RegistrationNumber
		profile.MerchantBusinessProfileTaxID = requestBody.TaxID
		profile.MerchantBusinessProfileWebsiteURL = requestBody.WebsiteURL
		profile.MerchantBusinessProfileIncoporationDate = requestBody.IncoporationDate
		profile.MerchantBusinessProfileLegalStructure = requestBody.LegalStructure
		profile.MerchantBusinessProfileMCCCode = requestBody.MCCCode
		profile.MerchantBusinessProfilePhoneNumber = requestBody.PhoneNumber
		profile.MerchantBusinessProfileEmail = requestBody.Email

		if err := h.DB.Save(&profile).Error; err != nil {
			http.Error(w, "Failed to update merchant business profile", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"merchant_business_profile_id": profile.MerchantBusinessProfileID,
			"message":                      "Merchant business profile updated",
		})
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "Failed to check existing merchant business profile", http.StatusInternalServerError)
		return
	}

	profile = models.MerchantBusinessProfile{
		MerchantBusinessProfileID:                 uuid.New().String(),
		MerchantBusinessProfileMerchantID:         requestBody.MerchantID,
		MerchantBusinessProfileDBAName:            requestBody.DBAName,
		MerchantBusinessProfileRegistrationNumber: requestBody.RegistrationNumber,
		MerchantBusinessProfileTaxID:              requestBody.TaxID,
		MerchantBusinessProfileWebsiteURL:         requestBody.WebsiteURL,
		MerchantBusinessProfileIncoporationDate:   requestBody.IncoporationDate,
		MerchantBusinessProfileLegalStructure:     requestBody.LegalStructure,
		MerchantBusinessProfileMCCCode:            requestBody.MCCCode,
		MerchantBusinessProfilePhoneNumber:        requestBody.PhoneNumber,
		MerchantBusinessProfileEmail:              requestBody.Email,
	}

	if err := h.DB.Create(&profile).Error; err != nil {
		http.Error(w, "Failed to create merchant business profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"merchant_business_profile_id": profile.MerchantBusinessProfileID,
		"message":                      "Merchant business profile created",
	})
}

func (h *Handler) CreateMerchantAddressHandler(w http.ResponseWriter, r *http.Request) {
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

	type CreateMerchantAddressRequest struct {
		MerchantID string `json:"merchant_id"`
		Line1      string `json:"line_1"`
		Line2      string `json:"line_2"`
		City       string `json:"city"`
		State      string `json:"state"`
		PostalCode int    `json:"postal_code"`
		Verified   bool   `json:"verified"`
	}

	var requestBody CreateMerchantAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}

	if requestBody.MerchantID == "" || requestBody.Line1 == "" || requestBody.Line2 == "" || requestBody.City == "" || requestBody.State == "" || requestBody.PostalCode == 0 {
		http.Error(w, "merchant_id, line_1, line_2, city, state, and postal_code are required", http.StatusBadRequest)
		return
	}

	if err := h.requireMerchantAccess(requestBody.MerchantID, sessionData.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to verify merchant and permissions", http.StatusInternalServerError)
		return
	}

	var address models.MerchantAddress
	err = h.DB.Where("merchant_address_merchant_id = ?", requestBody.MerchantID).First(&address).Error
	if err == nil {
		address.MerchantAddressLine1 = requestBody.Line1
		address.MerchantAddressLine2 = requestBody.Line2
		address.MerchantAddressCity = requestBody.City
		address.MerchantAddressState = requestBody.State
		address.MerchantAddressPostalCode = requestBody.PostalCode
		address.MerchantAddressVerified = requestBody.Verified

		if err := h.DB.Save(&address).Error; err != nil {
			http.Error(w, "Failed to update merchant address", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"merchant_address_id": address.MerchantAddressID,
			"message":             "Merchant address updated",
		})
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "Failed to check existing merchant address", http.StatusInternalServerError)
		return
	}

	address = models.MerchantAddress{
		MerchantAddressID:         uuid.New().String(),
		MerchantAddressMerchantID: requestBody.MerchantID,
		MerchantAddressLine1:      requestBody.Line1,
		MerchantAddressLine2:      requestBody.Line2,
		MerchantAddressCity:       requestBody.City,
		MerchantAddressState:      requestBody.State,
		MerchantAddressPostalCode: requestBody.PostalCode,
		MerchantAddressVerified:   requestBody.Verified,
	}

	if err := h.DB.Create(&address).Error; err != nil {
		http.Error(w, "Failed to create merchant address", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"merchant_address_id": address.MerchantAddressID,
		"message":             "Merchant address created",
	})
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
		EditorID   string `json:"editor_id"`
	}

	// try to parse request body
	var requestBody EditUserRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}
	// make sure request has all required params
	if requestBody.MerchantID == "" || requestBody.UserID == "" || requestBody.Role == "" || requestBody.EditorID == "" {
		http.Error(w, "merchant_id, user_id, and role are required", http.StatusBadRequest)
		return
	}
	// get user who is trying to make an edit's role
	var editorRole models.Role
	if err := h.DB.Where("role_user_id = ? AND role_merchant_id = ?", requestBody.EditorID, requestBody.MerchantID).First(&editorRole).Error; err != nil {
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

	type RemoveUserRequest struct {
		MerchantID string `json:"merchant_id"`
		UserID     string `json:"user_id"`
		EditorID   string `json:"editor_id"`
	}

	// try to parse request body
	var requestBody RemoveUserRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}

	// make sure request has all required params
	if requestBody.MerchantID == "" || requestBody.UserID == "" || requestBody.EditorID == "" {
		http.Error(w, "merchant_id, user_id, and editor_id are required", http.StatusBadRequest)
		return
	}
	// get user who is trying to remove's role
	var removerRole models.Role
	if err := h.DB.Where("role_user_id = ? AND role_merchant_id = ?", requestBody.EditorID, requestBody.MerchantID).First(&removerRole).Error; err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	// check if that role is allowed to edit
	if !rolesAllowedToEdit[removerRole.RoleName] {
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
	// get user role we're trying to edit
	var userRole models.Role
	if err := h.DB.Where("role_user_id = ?", requestBody.UserID).First(&userRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "No user found", http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to lookup user", http.StatusInternalServerError)
		return
	}
	// make the deletion of that role record
	h.DB.Delete(&userRole)

	// craft response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User removed",
	})
}

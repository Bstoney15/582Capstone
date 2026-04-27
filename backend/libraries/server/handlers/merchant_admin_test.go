package routes

// Tests covering the session-authenticated merchant administration handlers
// in merchant.go and the API-key management handlers in merchant_api_key.go.

import (
	"backend/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

// ---------- CreateMerchantHandler ----------

func TestCreateMerchantHandler_Unauthorized(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/create", bytes.NewReader([]byte("{}")))
	rec := callDirect(h.CreateMerchantHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCreateMerchantHandler_MissingFields(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "owner", "pw")
	body, _ := json.Marshal(map[string]any{"merchant_name": "Acme"})
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/create", bytes.NewReader(body))
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.CreateMerchantHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateMerchantHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "owner", "pw")
	payload := map[string]any{
		"merchant_name": "Acme",
		"merchant_business_profile_dba_name":            "Acme Co",
		"merchant_business_profile_registration_number": "REG1",
		"merchant_business_profile_tax_id":              "TAX1",
		"merchant_business_profile_website_url":         "https://a.com",
		"merchant_business_profile_incoporation_date":   "2020-01-01",
		"merchant_business_profile_legal_structure":     "LLC",
		"merchant_business_profile_mcc_code":            "5045",
		"merchant_business_profile_phone_number":        "555-1234",
		"merchant_business_profile_email":               "biz@a.com",
		"merchant_address_line_1":                       "1 Main",
		"merchant_address_line_2":                       "Apt 1",
		"merchant_address_city":                         "City",
		"merchant_address_state":                        "ST",
		"merchant_address_postal_code":                  12345,
		"merchant_owner_first_name":                     "O",
		"merchant_owner_last_name":                      "W",
		"merchant_owner_phone_number":                   "555-9999",
		"merchant_owner_email":                          "o@a.com",
		"merchant_owner_dob":                            "1990-01-01",
		"merchant_owner_stake":                          "100.0",
	}
	body, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/create", bytes.NewReader(body))
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.CreateMerchantHandler, r)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateMerchantHandler_BadDOB(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "owner", "pw")
	payload := map[string]any{
		"merchant_name":                                 "Acme",
		"merchant_business_profile_dba_name":            "x",
		"merchant_business_profile_registration_number": "x",
		"merchant_business_profile_tax_id":              "x",
		"merchant_business_profile_website_url":         "x",
		"merchant_business_profile_incoporation_date":   "x",
		"merchant_business_profile_legal_structure":     "x",
		"merchant_business_profile_mcc_code":            "x",
		"merchant_business_profile_phone_number":        "x",
		"merchant_business_profile_email":               "x",
		"merchant_address_line_1":                       "x",
		"merchant_address_line_2":                       "x",
		"merchant_address_city":                         "x",
		"merchant_address_state":                        "x",
		"merchant_address_postal_code":                  1,
		"merchant_owner_first_name":                     "x",
		"merchant_owner_last_name":                      "x",
		"merchant_owner_phone_number":                   "x",
		"merchant_owner_email":                          "x",
		"merchant_owner_dob":                            "not-a-date",
		"merchant_owner_stake":                          "1.0",
	}
	body, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/create", bytes.NewReader(body))
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.CreateMerchantHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// ---------- AddUserHandler ----------

func TestAddUserHandler_Unauthorized(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/add-user", bytes.NewReader([]byte("{}")))
	rec := callDirect(h.AddUserHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAddUserHandler_Forbidden_NotAdminOrOwner(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "dev", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleDeveloper)
	target := seedUser(t, h.DB, "target", "pw")

	body, _ := json.Marshal(map[string]string{
		"merchant_id": merchant.MerchantID, "user_username": target.UserUsername, "role": models.RoleDeveloper,
	})
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/add-user", bytes.NewReader(body))
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.AddUserHandler, r)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAddUserHandler_InvalidRole(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "admin", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)

	body, _ := json.Marshal(map[string]string{
		"merchant_id": merchant.MerchantID, "user_username": "x", "role": "Hacker",
	})
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/add-user", bytes.NewReader(body))
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.AddUserHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAddUserHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	admin := seedUser(t, h.DB, "admin", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, admin.UserID, merchant.MerchantID, models.RoleAdmin)
	target := seedUser(t, h.DB, "target", "pw")

	body, _ := json.Marshal(map[string]string{
		"merchant_id": merchant.MerchantID, "user_username": target.UserUsername, "role": models.RoleDeveloper,
	})
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/add-user", bytes.NewReader(body))
	r.AddCookie(loginAs(admin.UserID))
	rec := callDirect(h.AddUserHandler, r)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAddUserHandler_DuplicateRole(t *testing.T) {
	h := newTestHandler(t)
	admin := seedUser(t, h.DB, "admin", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, admin.UserID, merchant.MerchantID, models.RoleAdmin)
	target := seedUser(t, h.DB, "target", "pw")
	seedRole(t, h.DB, target.UserID, merchant.MerchantID, models.RoleDeveloper)

	body, _ := json.Marshal(map[string]string{
		"merchant_id": merchant.MerchantID, "user_username": target.UserUsername, "role": models.RoleDeveloper,
	})
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/add-user", bytes.NewReader(body))
	r.AddCookie(loginAs(admin.UserID))
	rec := callDirect(h.AddUserHandler, r)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

// ---------- EditUserHandler ----------

func TestEditUserHandler_Unauthorized(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodPatch, "/api/merchant/edit-user-role", bytes.NewReader([]byte("{}")))
	rec := callDirect(h.EditUserHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestEditUserHandler_MissingFields(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	body, _ := json.Marshal(map[string]string{"merchant_id": "x"})
	r := httptest.NewRequest(http.MethodPatch, "/api/merchant/edit-user-role", bytes.NewReader(body))
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.EditUserHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestEditUserHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	admin := seedUser(t, h.DB, "admin", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, admin.UserID, merchant.MerchantID, models.RoleAdmin)
	target := seedUser(t, h.DB, "target", "pw")
	seedRole(t, h.DB, target.UserID, merchant.MerchantID, models.RoleDeveloper)

	body, _ := json.Marshal(map[string]string{
		"merchant_id": merchant.MerchantID,
		"user_id":     target.UserID,
		"role":        models.RoleAdmin,
		"editor_id":   admin.UserID,
	})
	r := httptest.NewRequest(http.MethodPatch, "/api/merchant/edit-user-role", bytes.NewReader(body))
	r.AddCookie(loginAs(admin.UserID))
	rec := callDirect(h.EditUserHandler, r)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ---------- GetAllMerchantUsersHandler ----------

func TestGetAllMerchantUsersHandler_MissingMerchantID(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	r := httptest.NewRequest(http.MethodGet, "/api/merchant/get-merchant-users", nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetAllMerchantUsersHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetAllMerchantUsersHandler_NotMember(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	r := httptest.NewRequest(http.MethodGet, "/api/merchant/get-merchant-users?merchant_id=missing", nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetAllMerchantUsersHandler, r)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetAllMerchantUsersHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	admin := seedUser(t, h.DB, "admin", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, admin.UserID, merchant.MerchantID, models.RoleAdmin)

	r := httptest.NewRequest(http.MethodGet, "/api/merchant/get-merchant-users?merchant_id="+merchant.MerchantID, nil)
	r.AddCookie(loginAs(admin.UserID))
	rec := callDirect(h.GetAllMerchantUsersHandler, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ---------- RemoveMerchantUserHandler ----------

func TestRemoveMerchantUserHandler_Unauthorized(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodDelete, "/api/merchant/remove-user", bytes.NewReader([]byte("{}")))
	rec := callDirect(h.RemoveMerchantUserHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRemoveMerchantUserHandler_Forbidden(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	body, _ := json.Marshal(map[string]string{"merchant_id": "m", "user_id": "u", "editor_id": "e"})
	r := httptest.NewRequest(http.MethodDelete, "/api/merchant/remove-user", bytes.NewReader(body))
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.RemoveMerchantUserHandler, r)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRemoveMerchantUserHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	admin := seedUser(t, h.DB, "admin", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, admin.UserID, merchant.MerchantID, models.RoleAdmin)
	target := seedUser(t, h.DB, "target", "pw")
	seedRole(t, h.DB, target.UserID, merchant.MerchantID, models.RoleDeveloper)

	body, _ := json.Marshal(map[string]string{
		"merchant_id": merchant.MerchantID,
		"user_id":     target.UserID,
		"editor_id":   admin.UserID,
	})
	r := httptest.NewRequest(http.MethodDelete, "/api/merchant/remove-user", bytes.NewReader(body))
	r.AddCookie(loginAs(admin.UserID))
	rec := callDirect(h.RemoveMerchantUserHandler, r)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ---------- GetMerchantWalletHandler ----------

func TestGetMerchantWalletHandler_MissingMerchantID(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	r := httptest.NewRequest(http.MethodGet, "/api/merchant/get-wallet", nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetMerchantWalletHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetMerchantWalletHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	admin := seedUser(t, h.DB, "admin", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, admin.UserID, merchant.MerchantID, models.RoleAdmin)
	wallet := models.MerchantCryptoWallet{
		MerchantCryptoWalletID:         uuid.New().String(),
		MerchantCryptoWalletMerchantID: merchant.MerchantID,
		MerchantCryptoWalletAddress:    "rTestAddress",
	}
	if err := h.DB.Create(&wallet).Error; err != nil {
		t.Fatalf("seed wallet: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/api/merchant/get-wallet?merchant_id="+merchant.MerchantID, nil)
	r.AddCookie(loginAs(admin.UserID))
	rec := callDirect(h.GetMerchantWalletHandler, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ---------- SetMerchantWalletHandler ----------

func TestSetMerchantWalletHandler_NotAdminOwner(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "dev", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleDeveloper)

	body, _ := json.Marshal(map[string]string{"wallet_address": "rNew"})
	r := httptest.NewRequest(http.MethodPatch, "/api/merchant/set-wallet?merchant_id="+merchant.MerchantID, bytes.NewReader(body))
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.SetMerchantWalletHandler, r)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestSetMerchantWalletHandler_MissingBody(t *testing.T) {
	h := newTestHandler(t)
	admin := seedUser(t, h.DB, "admin", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, admin.UserID, merchant.MerchantID, models.RoleAdmin)

	body, _ := json.Marshal(map[string]string{})
	r := httptest.NewRequest(http.MethodPatch, "/api/merchant/set-wallet?merchant_id="+merchant.MerchantID, bytes.NewReader(body))
	r.AddCookie(loginAs(admin.UserID))
	rec := callDirect(h.SetMerchantWalletHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestSetMerchantWalletHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	admin := seedUser(t, h.DB, "admin", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, admin.UserID, merchant.MerchantID, models.RoleAdmin)
	wallet := models.MerchantCryptoWallet{
		MerchantCryptoWalletID:         uuid.New().String(),
		MerchantCryptoWalletMerchantID: merchant.MerchantID,
		MerchantCryptoWalletAddress:    "rOld",
	}
	if err := h.DB.Create(&wallet).Error; err != nil {
		t.Fatalf("seed wallet: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"wallet_address": "rNewAddr"})
	r := httptest.NewRequest(http.MethodPatch, "/api/merchant/set-wallet?merchant_id="+merchant.MerchantID, bytes.NewReader(body))
	r.AddCookie(loginAs(admin.UserID))
	rec := callDirect(h.SetMerchantWalletHandler, r)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ---------- GetMerchantCustomersHandler (session auth, dashboard view) ----------

func TestGetMerchantCustomersHandler_MissingMerchantID(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	r := httptest.NewRequest(http.MethodGet, "/api/customer", nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetMerchantCustomersHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetMerchantCustomersHandler_Forbidden(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	r := httptest.NewRequest(http.MethodGet, "/api/customer?merchant_id=mid", nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetMerchantCustomersHandler, r)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestGetMerchantCustomersHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleDeveloper)
	seedCustomer(t, h.DB, merchant.MerchantID)

	r := httptest.NewRequest(http.MethodGet, "/api/customer?merchant_id="+merchant.MerchantID, nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetMerchantCustomersHandler, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1 customer, got %d", len(resp))
	}
}

// ---------- Merchant API key handlers (session auth) ----------

func TestGetMerchantAPIKeysHandler_MissingMerchantID(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/merchant/api_key", nil)
	rec := callDirect(h.GetMerchantAPIKeysHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetMerchantAPIKeysHandler_Forbidden(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	r := httptest.NewRequest(http.MethodGet, "/api/merchant/api_key?merchant_id=foo", nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetMerchantAPIKeysHandler, r)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestGetMerchantAPIKeysHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)
	seedMerchantAPIKey(t, h.DB, merchant.MerchantID)

	r := httptest.NewRequest(http.MethodGet, "/api/merchant/api_key?merchant_id="+merchant.MerchantID, nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetMerchantAPIKeysHandler, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCreateMerchantAPIKeyHandler_BadJSON(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/api_key", bytes.NewBufferString("{"))
	rec := callDirect(h.CreateMerchantAPIKeyHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateMerchantAPIKeyHandler_MissingFields(t *testing.T) {
	h := newTestHandler(t)
	body, _ := json.Marshal(createMerchantAPIKeyRequest{})
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/api_key", bytes.NewReader(body))
	rec := callDirect(h.CreateMerchantAPIKeyHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateMerchantAPIKeyHandler_Forbidden(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	body, _ := json.Marshal(createMerchantAPIKeyRequest{MerchantID: "missing", Name: "key1"})
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/api_key", bytes.NewReader(body))
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.CreateMerchantAPIKeyHandler, r)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestCreateMerchantAPIKeyHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)

	body, _ := json.Marshal(createMerchantAPIKeyRequest{MerchantID: merchant.MerchantID, Name: "key1"})
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/api_key", bytes.NewReader(body))
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.CreateMerchantAPIKeyHandler, r)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp createMerchantAPIKeyResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.APIKey == "" {
		t.Errorf("expected api_key in response")
	}
}

func TestDeleteMerchantAPIKeyHandler_MissingPath(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodDelete, "/api/merchant/api_key/", nil)
	r.SetPathValue("api_key", "")
	rec := callDirect(h.DeleteMerchantAPIKeyHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDeleteMerchantAPIKeyHandler_NotFound(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)

	r := httptest.NewRequest(http.MethodDelete, "/api/merchant/api_key/missing?merchant_id="+merchant.MerchantID, nil)
	r.SetPathValue("api_key", "missing")
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.DeleteMerchantAPIKeyHandler, r)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteMerchantAPIKeyHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)

	keyID := uuid.New().String()
	if err := h.DB.Create(&models.MerchantAPIKey{
		MerchantAPIKeyID: keyID, MerchantAPIKeyName: "n",
		MerchantAPIKeyHashed: "deadbeef", MerchantAPIKeyMerchantID: merchant.MerchantID,
	}).Error; err != nil {
		t.Fatalf("seed key: %v", err)
	}

	r := httptest.NewRequest(http.MethodDelete, "/api/merchant/api_key/"+keyID+"?merchant_id="+merchant.MerchantID, nil)
	r.SetPathValue("api_key", keyID)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.DeleteMerchantAPIKeyHandler, r)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	var stored models.MerchantAPIKey
	h.DB.Where("merchant_api_key_id = ?", keyID).First(&stored)
	if !stored.MerchantAPIKeyRevoked {
		t.Errorf("expected key to be marked revoked")
	}
}

package routes

// Tests covering the authentication and user-info handlers:
//   login.go, logout.go, signup.go, auth.go (CheckAuthHandler),
//   user_info.go (GetUserInfo), merchant.go (GetMerchantsHandler),
//   settings.go (LeaveMerchantHandler).

import (
	"backend/libraries/sessionManager"
	"backend/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---------- LoginHandler ----------

func TestLoginHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	seedUser(t, h.DB, "alice", "secret123")

	body, _ := json.Marshal(LoginRequest{Username: "alice", Password: "secret123"})
	r := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	rec := callDirect(h.LoginHandler, r)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == sessionManager.SessionCookieName && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected session cookie to be set")
	}
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader("{"))
	rec := callDirect(h.LoginHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestLoginHandler_UnknownUser(t *testing.T) {
	h := newTestHandler(t)
	body, _ := json.Marshal(LoginRequest{Username: "nobody", Password: "x"})
	r := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	rec := callDirect(h.LoginHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestLoginHandler_WrongPassword(t *testing.T) {
	h := newTestHandler(t)
	seedUser(t, h.DB, "bob", "correct-password")

	body, _ := json.Marshal(LoginRequest{Username: "bob", Password: "wrong"})
	r := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	rec := callDirect(h.LoginHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

// ---------- LogoutHandler ----------

func TestLogoutHandler_ClearsCookie(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodPost, "/api/user/logout", nil)
	rec := callDirect(h.LogoutHandler, r)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	cookies := rec.Result().Cookies()
	cleared := false
	for _, c := range cookies {
		if c.Name == sessionManager.SessionCookieName && c.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Errorf("expected session cookie to be cleared")
	}
}

// ---------- SignupHandler ----------

func TestSignupHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	body, _ := json.Marshal(SignupRequest{
		FirstName: "New", LastName: "User", Username: "newuser", Password: "password",
	})
	r := httptest.NewRequest(http.MethodPost, "/api/user/signup", bytes.NewReader(body))
	rec := callDirect(h.SignupHandler, r)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var stored models.User
	if err := h.DB.Where("user_username = ?", "newuser").First(&stored).Error; err != nil {
		t.Fatalf("user not persisted: %v", err)
	}
	if stored.UserPasswordHash == "password" || stored.UserPasswordHash == "" {
		t.Errorf("password should be hashed, got %q", stored.UserPasswordHash)
	}
}

func TestSignupHandler_DuplicateUsername(t *testing.T) {
	h := newTestHandler(t)
	seedUser(t, h.DB, "taken", "pw")

	body, _ := json.Marshal(SignupRequest{Username: "taken", Password: "pw", FirstName: "F", LastName: "L"})
	r := httptest.NewRequest(http.MethodPost, "/api/user/signup", bytes.NewReader(body))
	rec := callDirect(h.SignupHandler, r)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestSignupHandler_InvalidJSON(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodPost, "/api/user/signup", strings.NewReader("{"))
	rec := callDirect(h.SignupHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// ---------- CheckAuthHandler ----------

func TestCheckAuthHandler_NoCookie(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/user/auth", nil)
	rec := callDirect(h.CheckAuthHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCheckAuthHandler_InvalidToken(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/user/auth", nil)
	r.AddCookie(&http.Cookie{Name: sessionManager.SessionCookieName, Value: "garbage"})
	rec := callDirect(h.CheckAuthHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCheckAuthHandler_Authenticated(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "auth-user", "x")

	r := httptest.NewRequest(http.MethodGet, "/api/user/auth", nil)
	r.AddCookie(loginAs(user.UserID))

	rec := callDirect(h.CheckAuthHandler, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.Authenticated || resp.UserID != user.UserID {
		t.Errorf("unexpected response: %+v", resp)
	}
}

// ---------- GetUserInfo ----------

func TestGetUserInfo_Unauthorized(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/user/info", nil)
	rec := callDirect(h.GetUserInfo, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestGetUserInfo_Success(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "info-user", "pw")

	r := httptest.NewRequest(http.MethodGet, "/api/user/info", nil)
	r.AddCookie(loginAs(user.UserID))

	rec := callDirect(h.GetUserInfo, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp UserProfileResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.UserID != user.UserID || resp.Username != "info-user" {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestGetUserInfo_UserDeleted(t *testing.T) {
	h := newTestHandler(t)
	// Session for a user id that does not exist in the database.
	r := httptest.NewRequest(http.MethodGet, "/api/user/info", nil)
	r.AddCookie(loginAs("ghost-user-id"))

	rec := callDirect(h.GetUserInfo, r)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// ---------- GetMerchantsHandler ----------

func TestGetMerchantsHandler_Unauthorized(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/user/merchants", nil)
	rec := callDirect(h.GetMerchantsHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestGetMerchantsHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "p")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)

	r := httptest.NewRequest(http.MethodGet, "/api/user/merchants", nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetMerchantsHandler, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp []MerchantResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp) != 1 || resp[0].ID != merchant.MerchantID || resp[0].Role != models.RoleAdmin {
		t.Errorf("unexpected: %+v", resp)
	}
}

// ---------- LeaveMerchantHandler ----------

func TestLeaveMerchantHandler_Unauthorized(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodDelete, "/api/user/merchants/abc", nil)
	r.SetPathValue("merchant_id", "abc")
	rec := callDirect(h.LeaveMerchantHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestLeaveMerchantHandler_MissingPath(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "leaver", "pw")
	r := httptest.NewRequest(http.MethodDelete, "/api/user/merchants/", nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.LeaveMerchantHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestLeaveMerchantHandler_NotAssociated(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "leaver", "pw")
	r := httptest.NewRequest(http.MethodDelete, "/api/user/merchants/missing", nil)
	r.SetPathValue("merchant_id", "missing")
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.LeaveMerchantHandler, r)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestLeaveMerchantHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "leaver", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)

	r := httptest.NewRequest(http.MethodDelete, "/api/user/merchants/"+merchant.MerchantID, nil)
	r.SetPathValue("merchant_id", merchant.MerchantID)
	r.AddCookie(loginAs(user.UserID))

	rec := callDirect(h.LeaveMerchantHandler, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var count int64
	h.DB.Model(&models.Role{}).Where("role_user_id = ?", user.UserID).Count(&count)
	if count != 0 {
		t.Errorf("role not deleted, count=%d", count)
	}
}

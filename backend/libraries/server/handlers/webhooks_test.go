package routes

// Tests for the merchant webhook config handlers (merchant_webhook.go) and
// webhook log handlers (webhook_log.go).

import (
	"backend/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

// ---------- GetMerchantWebhooksHandler ----------

func TestGetMerchantWebhooksHandler_MissingMerchantID(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/merchant/webhooks", nil)
	rec := callDirect(h.GetMerchantWebhooksHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetMerchantWebhooksHandler_Forbidden(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	r := httptest.NewRequest(http.MethodGet, "/api/merchant/webhooks?merchant_id=foo", nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetMerchantWebhooksHandler, r)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestGetMerchantWebhooksHandler_Success_Empty(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)

	r := httptest.NewRequest(http.MethodGet, "/api/merchant/webhooks?merchant_id="+merchant.MerchantID, nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetMerchantWebhooksHandler, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp []webhookSummary
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty list, got %d", len(resp))
	}
}

// ---------- CreateMerchantWebhookHandler ----------

func TestCreateMerchantWebhookHandler_BadJSON(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/webhooks", bytes.NewBufferString("{"))
	rec := callDirect(h.CreateMerchantWebhookHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateMerchantWebhookHandler_MissingFields(t *testing.T) {
	h := newTestHandler(t)
	body, _ := json.Marshal(map[string]string{"merchant_id": "x"})
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/webhooks", bytes.NewReader(body))
	rec := callDirect(h.CreateMerchantWebhookHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateMerchantWebhookHandler_Forbidden(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	body, _ := json.Marshal(map[string]string{
		"merchant_id": "missing", "url": "https://x", "secret": "s",
	})
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/webhooks", bytes.NewReader(body))
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.CreateMerchantWebhookHandler, r)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestCreateMerchantWebhookHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)

	body, _ := json.Marshal(map[string]string{
		"merchant_id": merchant.MerchantID, "url": "https://hook.example", "secret": "shh",
	})
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/webhooks", bytes.NewReader(body))
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.CreateMerchantWebhookHandler, r)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateMerchantWebhookHandler_UpdateExisting(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)
	if err := h.DB.Create(&models.MerchantWebhookKey{
		MerchantWebhookKeyID: uuid.New().String(),
		MerchantWebhookURL:   "https://old", MerchantWebhookKey: "old",
		MerchantWebhookKeyMerchantID: merchant.MerchantID,
	}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	body, _ := json.Marshal(map[string]string{
		"merchant_id": merchant.MerchantID, "url": "https://new", "secret": "new",
	})
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/webhooks", bytes.NewReader(body))
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.CreateMerchantWebhookHandler, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var stored models.MerchantWebhookKey
	h.DB.Where("merchant_webhook_key_merchant_id = ?", merchant.MerchantID).First(&stored)
	if stored.MerchantWebhookURL != "https://new" {
		t.Errorf("expected URL update, got %q", stored.MerchantWebhookURL)
	}
}

// ---------- DeleteMerchantWebhookHandler ----------

func TestDeleteMerchantWebhookHandler_MissingPath(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodDelete, "/api/merchant/webhooks/", nil)
	r.SetPathValue("webhook_id", "")
	rec := callDirect(h.DeleteMerchantWebhookHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDeleteMerchantWebhookHandler_NotFound(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)

	r := httptest.NewRequest(http.MethodDelete, "/api/merchant/webhooks/missing?merchant_id="+merchant.MerchantID, nil)
	r.SetPathValue("webhook_id", "missing")
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.DeleteMerchantWebhookHandler, r)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteMerchantWebhookHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)

	hookID := uuid.New().String()
	if err := h.DB.Create(&models.MerchantWebhookKey{
		MerchantWebhookKeyID:         hookID,
		MerchantWebhookURL:           "https://x",
		MerchantWebhookKey:           "k",
		MerchantWebhookKeyMerchantID: merchant.MerchantID,
	}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	r := httptest.NewRequest(http.MethodDelete, "/api/merchant/webhooks/"+hookID+"?merchant_id="+merchant.MerchantID, nil)
	r.SetPathValue("webhook_id", hookID)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.DeleteMerchantWebhookHandler, r)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

// ---------- ListWebhookLogsHandler ----------

func TestListWebhookLogsHandler_MissingMerchantID(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/merchant/webhook_logs", nil)
	rec := callDirect(h.ListWebhookLogsHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestListWebhookLogsHandler_Forbidden(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	r := httptest.NewRequest(http.MethodGet, "/api/merchant/webhook_logs?merchant_id=foo", nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.ListWebhookLogsHandler, r)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestListWebhookLogsHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)
	if err := h.DB.Create(&models.WebhookLog{
		WebhookLogID:         uuid.New().String(),
		WebhookLogMerchantID: merchant.MerchantID,
		WebhookLogEventType:  "invoice.paid",
		WebhookLogPayload:    "{}",
		WebhookLogSucceeded:  true,
	}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/api/merchant/webhook_logs?merchant_id="+merchant.MerchantID, nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.ListWebhookLogsHandler, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp []webhookLogSummary
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1 log, got %d", len(resp))
	}
}

// ---------- ResendWebhookHandler ----------

func TestResendWebhookHandler_MissingParams(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodPost, "/api/merchant/webhook_logs//resend", nil)
	r.SetPathValue("log_id", "")
	rec := callDirect(h.ResendWebhookHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestResendWebhookHandler_LogNotFound(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)

	r := httptest.NewRequest(http.MethodPost, "/api/merchant/webhook_logs/missing/resend?merchant_id="+merchant.MerchantID, nil)
	r.SetPathValue("log_id", "missing")
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.ResendWebhookHandler, r)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestResendWebhookHandler_NoConfiguredWebhook(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)
	logID := uuid.New().String()
	if err := h.DB.Create(&models.WebhookLog{
		WebhookLogID:         logID,
		WebhookLogMerchantID: merchant.MerchantID,
		WebhookLogEventType:  "invoice.paid",
		WebhookLogPayload:    "{}",
	}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	r := httptest.NewRequest(http.MethodPost, "/api/merchant/webhook_logs/"+logID+"/resend?merchant_id="+merchant.MerchantID, nil)
	r.SetPathValue("log_id", logID)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.ResendWebhookHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestResendWebhookHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)

	// Stand up a fake receiver to accept the resend.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := h.DB.Create(&models.MerchantWebhookKey{
		MerchantWebhookKeyID:         uuid.New().String(),
		MerchantWebhookURL:           server.URL,
		MerchantWebhookKey:           "secret",
		MerchantWebhookKeyMerchantID: merchant.MerchantID,
	}).Error; err != nil {
		t.Fatalf("seed hook: %v", err)
	}

	logID := uuid.New().String()
	if err := h.DB.Create(&models.WebhookLog{
		WebhookLogID:         logID,
		WebhookLogMerchantID: merchant.MerchantID,
		WebhookLogEventType:  "invoice.paid",
		WebhookLogPayload:    `{"k":"v"}`,
	}).Error; err != nil {
		t.Fatalf("seed log: %v", err)
	}

	r := httptest.NewRequest(http.MethodPost, "/api/merchant/webhook_logs/"+logID+"/resend?merchant_id="+merchant.MerchantID, nil)
	r.SetPathValue("log_id", logID)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.ResendWebhookHandler, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

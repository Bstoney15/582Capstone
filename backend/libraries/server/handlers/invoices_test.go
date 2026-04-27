package routes

// Tests for the invoice and verify handlers:
//   invoice_customer.go (GetCustomerInvoicesHandler)
//   invoice_public.go   (GetInvoiceForCheckoutHandler)
//   invoice_events.go   (StreamInvoiceEventsHandler)
//   verify.go           (VerifyInvoicePaymentHandler)

import (
	"backend/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func mustSeedInvoice(t *testing.T, db any, customerID, status string) string {
	t.Helper()
	h, ok := db.(*Handler)
	if !ok {
		t.Fatalf("expected *Handler")
	}
	id := uuid.New().String()
	tag := uint32(123)
	if err := h.DB.Create(&models.Invoice{
		InvoiceID:             id,
		InvoiceAmountCharged:  decimal.NewFromFloat(2.5),
		InvoiceDestinationTag: &tag,
		InvoiceStatus:         status,
		InvoiceFeeAmount:      decimal.Zero,
		InvoiceFeeStatus:      "unpaid",
		InvoiceCryptoType:     "XRP",
		InvoiceCustomerID:     customerID,
	}).Error; err != nil {
		t.Fatalf("seed invoice: %v", err)
	}
	return id
}

// ---------- GetCustomerInvoicesHandler ----------

func TestGetCustomerInvoicesHandler_Unauthorized(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/customer/x/invoices", nil)
	r.SetPathValue("customer_id", "x")
	rec := callDirect(h.GetCustomerInvoicesHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestGetCustomerInvoicesHandler_MissingPath(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	r := httptest.NewRequest(http.MethodGet, "/api/customer//invoices", nil)
	r.SetPathValue("customer_id", "")
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetCustomerInvoicesHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetCustomerInvoicesHandler_CustomerNotFound(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	r := httptest.NewRequest(http.MethodGet, "/api/customer/missing/invoices", nil)
	r.SetPathValue("customer_id", "missing")
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetCustomerInvoicesHandler, r)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetCustomerInvoicesHandler_Forbidden(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)
	// no role for user on this merchant

	r := httptest.NewRequest(http.MethodGet, "/api/customer/"+customer.CustomerID+"/invoices", nil)
	r.SetPathValue("customer_id", customer.CustomerID)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetCustomerInvoicesHandler, r)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestGetCustomerInvoicesHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)
	mustSeedInvoice(t, h, customer.CustomerID, "created")

	r := httptest.NewRequest(http.MethodGet, "/api/customer/"+customer.CustomerID+"/invoices", nil)
	r.SetPathValue("customer_id", customer.CustomerID)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetCustomerInvoicesHandler, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp []CustomerInvoiceResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1, got %d", len(resp))
	}
}

// ---------- GetInvoiceForCheckoutHandler ----------

func TestGetInvoiceForCheckoutHandler_MissingPath(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/invoices/", nil)
	r.SetPathValue("uuid", "")
	rec := callDirect(h.GetInvoiceForCheckoutHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetInvoiceForCheckoutHandler_InvoiceNotFound(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/invoices/missing", nil)
	r.SetPathValue("uuid", "missing")
	rec := callDirect(h.GetInvoiceForCheckoutHandler, r)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetInvoiceForCheckoutHandler_NoVerifiedWallet(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)
	id := mustSeedInvoice(t, h, customer.CustomerID, "created")

	r := httptest.NewRequest(http.MethodGet, "/api/invoices/"+id, nil)
	r.SetPathValue("uuid", id)
	rec := callDirect(h.GetInvoiceForCheckoutHandler, r)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetInvoiceForCheckoutHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)
	id := mustSeedInvoice(t, h, customer.CustomerID, "created")
	if err := h.DB.Create(&models.MerchantCryptoWallet{
		MerchantCryptoWalletID:         uuid.New().String(),
		MerchantCryptoWalletMerchantID: merchant.MerchantID,
		MerchantCryptoWalletAddress:    "rDestAddr",
		MerchantCryptoWalletVerified:   true,
	}).Error; err != nil {
		t.Fatalf("seed wallet: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/api/invoices/"+id, nil)
	r.SetPathValue("uuid", id)
	rec := callDirect(h.GetInvoiceForCheckoutHandler, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp InvoiceCheckoutResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.MerchantAddress != "rDestAddr" {
		t.Errorf("merchant address = %q", resp.MerchantAddress)
	}
}

// ---------- StreamInvoiceEventsHandler ----------

// streamingRecorder is an httptest.ResponseRecorder that also implements http.Flusher.
type streamingRecorder struct {
	*httptest.ResponseRecorder
}

func (s *streamingRecorder) Flush() {}

func TestStreamInvoiceEventsHandler_MissingPath(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/invoices//events", nil)
	r.SetPathValue("uuid", "")
	rec := &streamingRecorder{httptest.NewRecorder()}
	h.StreamInvoiceEventsHandler(rec, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestStreamInvoiceEventsHandler_NotFound(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/invoices/missing/events", nil)
	r.SetPathValue("uuid", "missing")
	rec := &streamingRecorder{httptest.NewRecorder()}
	h.StreamInvoiceEventsHandler(rec, r)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestStreamInvoiceEventsHandler_TerminalStatusReturnsImmediately(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)
	id := mustSeedInvoice(t, h, customer.CustomerID, "paid")

	r := httptest.NewRequest(http.MethodGet, "/api/invoices/"+id+"/events", nil)
	r.SetPathValue("uuid", id)
	rec := &streamingRecorder{httptest.NewRecorder()}

	done := make(chan struct{})
	go func() {
		h.StreamInvoiceEventsHandler(rec, r)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not return for terminal status")
	}

	if !strings.Contains(rec.Body.String(), `"status":"paid"`) {
		t.Errorf("expected paid event, got: %s", rec.Body.String())
	}
}

// ---------- VerifyInvoicePaymentHandler ----------

func TestVerifyInvoicePaymentHandler_BadJSON(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodPost, "/api/verify", strings.NewReader("{"))
	rec := callDirect(h.VerifyInvoicePaymentHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestVerifyInvoicePaymentHandler_MissingFields(t *testing.T) {
	h := newTestHandler(t)
	body, _ := json.Marshal(VerifyInvoiceRequest{})
	r := httptest.NewRequest(http.MethodPost, "/api/verify", bytes.NewReader(body))
	rec := callDirect(h.VerifyInvoicePaymentHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestVerifyInvoicePaymentHandler_NotFound(t *testing.T) {
	h := newTestHandler(t)
	body, _ := json.Marshal(VerifyInvoiceRequest{InvoiceID: "missing", TxHash: "abc"})
	r := httptest.NewRequest(http.MethodPost, "/api/verify", bytes.NewReader(body))
	rec := callDirect(h.VerifyInvoicePaymentHandler, r)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestVerifyInvoicePaymentHandler_Accepted_ValidXRPLHash(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)
	id := mustSeedInvoice(t, h, customer.CustomerID, "created")

	hash := strings.Repeat("A", 64)
	body, _ := json.Marshal(VerifyInvoiceRequest{InvoiceID: id, TxHash: hash})
	r := httptest.NewRequest(http.MethodPost, "/api/verify", bytes.NewReader(body))
	rec := callDirect(h.VerifyInvoicePaymentHandler, r)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}

	// Allow background goroutine to update status.
	for i := 0; i < 50; i++ {
		var inv models.Invoice
		h.DB.Where("invoice_id = ?", id).First(&inv)
		if inv.InvoiceStatus == "paid" {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("invoice status was not updated to paid")
}

func TestVerifyInvoicePaymentHandler_Accepted_InvalidHashMarksFailed(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)
	id := mustSeedInvoice(t, h, customer.CustomerID, "created")

	body, _ := json.Marshal(VerifyInvoiceRequest{InvoiceID: id, TxHash: "not-a-real-hash"})
	r := httptest.NewRequest(http.MethodPost, "/api/verify", bytes.NewReader(body))
	rec := callDirect(h.VerifyInvoicePaymentHandler, r)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}

	for i := 0; i < 50; i++ {
		var inv models.Invoice
		h.DB.Where("invoice_id = ?", id).First(&inv)
		if inv.InvoiceStatus == "verification_failed" {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("invoice status was not updated to verification_failed")
}

package routes

import (
	"backend/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateInvoiceHandler_Success_XRP(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)
	seedCustomer(t, h.DB, merchant.MerchantID)

	body, _ := json.Marshal(CreateInvoiceRequest{AmountXRP: "12.5000"})
	r := httptest.NewRequest(http.MethodPost, "/api/invoices", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	rec := callWithAPIKey(h.DB, h.CreateInvoiceHandler, r, apiKey)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp CreateInvoiceResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.InvoiceID == "" {
		t.Errorf("missing invoice_id")
	}
	if resp.AmountXRP != "12.5000" {
		t.Errorf("amount_xrp = %q, want 12.5000", resp.AmountXRP)
	}
	if resp.PricingSource != "direct_xrp" {
		t.Errorf("pricing_source = %q, want direct_xrp", resp.PricingSource)
	}

	var stored models.Invoice
	if err := h.DB.First(&stored, "invoice_id = ?", resp.InvoiceID).Error; err != nil {
		t.Fatalf("invoice not persisted: %v", err)
	}
	if stored.InvoiceStatus != "created" {
		t.Errorf("status = %q, want created", stored.InvoiceStatus)
	}
	if stored.InvoiceDestinationTag == nil || *stored.InvoiceDestinationTag == 0 {
		t.Errorf("destination tag should be allocated and non-zero")
	}
}

func TestCreateInvoiceHandler_InvalidJSON(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)

	r := httptest.NewRequest(http.MethodPost, "/api/invoices", strings.NewReader("{not json"))
	r.Header.Set("Content-Type", "application/json")

	rec := callWithAPIKey(h.DB, h.CreateInvoiceHandler, r, apiKey)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateInvoiceHandler_MissingMerchantContext(t *testing.T) {
	h := newTestHandler(t)
	body, _ := json.Marshal(CreateInvoiceRequest{AmountXRP: "1.0"})
	r := httptest.NewRequest(http.MethodPost, "/api/invoices", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	// Bypass middleware → no merchant id in context.
	rec := callDirect(h.CreateInvoiceHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCreateInvoiceHandler_InvalidAPIKey(t *testing.T) {
	h := newTestHandler(t)
	body, _ := json.Marshal(CreateInvoiceRequest{AmountXRP: "1.0"})
	r := httptest.NewRequest(http.MethodPost, "/api/invoices", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	rec := callWithAPIKey(h.DB, h.CreateInvoiceHandler, r, "definitely-not-a-real-key")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCreateInvoiceHandler_NonPositiveXRP(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)
	seedCustomer(t, h.DB, merchant.MerchantID)

	body, _ := json.Marshal(CreateInvoiceRequest{AmountXRP: "0"})
	r := httptest.NewRequest(http.MethodPost, "/api/invoices", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	rec := callWithAPIKey(h.DB, h.CreateInvoiceHandler, r, apiKey)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateInvoiceHandler_UnparseableXRP(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)
	seedCustomer(t, h.DB, merchant.MerchantID)

	body, _ := json.Marshal(CreateInvoiceRequest{AmountXRP: "not-a-number"})
	r := httptest.NewRequest(http.MethodPost, "/api/invoices", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	rec := callWithAPIKey(h.DB, h.CreateInvoiceHandler, r, apiKey)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateInvoiceHandler_NoCustomerForMerchant(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)
	// no customer seeded

	body, _ := json.Marshal(CreateInvoiceRequest{AmountXRP: "5.0"})
	r := httptest.NewRequest(http.MethodPost, "/api/invoices", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	rec := callWithAPIKey(h.DB, h.CreateInvoiceHandler, r, apiKey)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateInvoiceHandler_NeitherAmountProvided(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)
	seedCustomer(t, h.DB, merchant.MerchantID)

	body, _ := json.Marshal(CreateInvoiceRequest{})
	r := httptest.NewRequest(http.MethodPost, "/api/invoices", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	rec := callWithAPIKey(h.DB, h.CreateInvoiceHandler, r, apiKey)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

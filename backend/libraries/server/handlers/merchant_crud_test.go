package routes

// Tests for the merchant API-key authenticated CRUD handlers in merchant_crud.go:
// customer create/list/get/update/delete and invoice create/list/get/delete.

import (
	"backend/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ---------- CreateMerchantCustomerHandler ----------

func TestCreateMerchantCustomerHandler_Unauthorized(t *testing.T) {
	h := newTestHandler(t)
	body, _ := json.Marshal(createCustomerRequest{FirstName: "f"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/merchant/customers", bytes.NewReader(body))
	rec := callDirect(h.CreateMerchantCustomerHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCreateMerchantCustomerHandler_BadJSON(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/merchant/customers", strings.NewReader("{"))
	r.Header.Set("Content-Type", "application/json")
	rec := callWithAPIKey(h.DB, h.CreateMerchantCustomerHandler, r, apiKey)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateMerchantCustomerHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)

	body, _ := json.Marshal(createCustomerRequest{FirstName: "Jane", LastName: "Doe", Email: "j@d.com"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/merchant/customers", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	rec := callWithAPIKey(h.DB, h.CreateMerchantCustomerHandler, r, apiKey)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ---------- ListMerchantCustomersHandler ----------

func TestListMerchantCustomersHandler_Unauthorized(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/customers", nil)
	rec := callDirect(h.ListMerchantCustomersHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestListMerchantCustomersHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)
	seedCustomer(t, h.DB, merchant.MerchantID)

	r := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/customers", nil)
	rec := callWithAPIKey(h.DB, h.ListMerchantCustomersHandler, r, apiKey)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var customers []models.Customer
	if err := json.NewDecoder(rec.Body).Decode(&customers); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(customers) != 1 {
		t.Errorf("expected 1, got %d", len(customers))
	}
}

// ---------- GetMerchantCustomerHandler ----------

func TestGetMerchantCustomerHandler_MissingPath(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)

	r := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/customers/", nil)
	r.SetPathValue("customer_id", "")
	rec := callWithAPIKey(h.DB, h.GetMerchantCustomerHandler, r, apiKey)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetMerchantCustomerHandler_NotFound(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)

	r := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/customers/missing", nil)
	r.SetPathValue("customer_id", "missing")
	rec := callWithAPIKey(h.DB, h.GetMerchantCustomerHandler, r, apiKey)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetMerchantCustomerHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)

	r := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/customers/"+customer.CustomerID, nil)
	r.SetPathValue("customer_id", customer.CustomerID)
	rec := callWithAPIKey(h.DB, h.GetMerchantCustomerHandler, r, apiKey)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// ---------- UpdateMerchantCustomerHandler ----------

func TestUpdateMerchantCustomerHandler_NoFields(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)

	body, _ := json.Marshal(updateCustomerRequest{})
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/merchant/customers/"+customer.CustomerID, bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.SetPathValue("customer_id", customer.CustomerID)
	rec := callWithAPIKey(h.DB, h.UpdateMerchantCustomerHandler, r, apiKey)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateMerchantCustomerHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)

	newName := "Renamed"
	body, _ := json.Marshal(updateCustomerRequest{FirstName: &newName})
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/merchant/customers/"+customer.CustomerID, bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.SetPathValue("customer_id", customer.CustomerID)
	rec := callWithAPIKey(h.DB, h.UpdateMerchantCustomerHandler, r, apiKey)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var updated models.Customer
	h.DB.Where("customer_id = ?", customer.CustomerID).First(&updated)
	if updated.CustomerFirstName != newName {
		t.Errorf("first name = %q, want %q", updated.CustomerFirstName, newName)
	}
}

func TestUpdateMerchantCustomerHandler_NotFound(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)

	newName := "X"
	body, _ := json.Marshal(updateCustomerRequest{FirstName: &newName})
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/merchant/customers/missing", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.SetPathValue("customer_id", "missing")
	rec := callWithAPIKey(h.DB, h.UpdateMerchantCustomerHandler, r, apiKey)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// ---------- DeleteMerchantCustomerHandler ----------

func TestDeleteMerchantCustomerHandler_NotFound(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)

	r := httptest.NewRequest(http.MethodDelete, "/api/v1/merchant/customers/missing", nil)
	r.SetPathValue("customer_id", "missing")
	rec := callWithAPIKey(h.DB, h.DeleteMerchantCustomerHandler, r, apiKey)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteMerchantCustomerHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)

	r := httptest.NewRequest(http.MethodDelete, "/api/v1/merchant/customers/"+customer.CustomerID, nil)
	r.SetPathValue("customer_id", customer.CustomerID)
	rec := callWithAPIKey(h.DB, h.DeleteMerchantCustomerHandler, r, apiKey)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

// ---------- CreateMerchantInvoiceHandler ----------

func TestCreateMerchantInvoiceHandler_MissingCustomer(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)

	body, _ := json.Marshal(createInvoiceMerchantRequest{AmountCharged: "10"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/merchant/invoices", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	rec := callWithAPIKey(h.DB, h.CreateMerchantInvoiceHandler, r, apiKey)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateMerchantInvoiceHandler_BadAmount(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)

	body, _ := json.Marshal(createInvoiceMerchantRequest{CustomerID: customer.CustomerID, AmountCharged: "0"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/merchant/invoices", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	rec := callWithAPIKey(h.DB, h.CreateMerchantInvoiceHandler, r, apiKey)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateMerchantInvoiceHandler_CustomerNotFound(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)

	body, _ := json.Marshal(createInvoiceMerchantRequest{CustomerID: "missing", AmountCharged: "10"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/merchant/invoices", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	rec := callWithAPIKey(h.DB, h.CreateMerchantInvoiceHandler, r, apiKey)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestCreateMerchantInvoiceHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)

	body, _ := json.Marshal(createInvoiceMerchantRequest{
		CustomerID: customer.CustomerID, AmountCharged: "25.50", FeeAmount: "0.50",
	})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/merchant/invoices", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	rec := callWithAPIKey(h.DB, h.CreateMerchantInvoiceHandler, r, apiKey)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ---------- ListMerchantInvoicesHandler ----------

func TestListMerchantInvoicesHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)
	if err := h.DB.Create(&models.Invoice{
		InvoiceID:            uuid.New().String(),
		InvoiceAmountCharged: decimal.NewFromInt(10),
		InvoiceStatus:        "paid",
		InvoiceFeeAmount:     decimal.Zero,
		InvoiceFeeStatus:     "unpaid",
		InvoiceCryptoType:    "XRP",
		InvoiceCustomerID:    customer.CustomerID,
	}).Error; err != nil {
		t.Fatalf("seed invoice: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/invoices?completed_only=true", nil)
	rec := callWithAPIKey(h.DB, h.ListMerchantInvoicesHandler, r, apiKey)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestListMerchantInvoicesHandler_Unauthorized(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/invoices", nil)
	rec := callDirect(h.ListMerchantInvoicesHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

// ---------- GetMerchantInvoiceHandler ----------

func TestGetMerchantInvoiceHandler_NotFound(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)

	r := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/invoices/missing", nil)
	r.SetPathValue("invoice_id", "missing")
	rec := callWithAPIKey(h.DB, h.GetMerchantInvoiceHandler, r, apiKey)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetMerchantInvoiceHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)
	invoiceID := uuid.New().String()
	if err := h.DB.Create(&models.Invoice{
		InvoiceID:            invoiceID,
		InvoiceAmountCharged: decimal.NewFromInt(10),
		InvoiceStatus:        "created",
		InvoiceFeeAmount:     decimal.Zero,
		InvoiceFeeStatus:     "unpaid",
		InvoiceCryptoType:    "XRP",
		InvoiceCustomerID:    customer.CustomerID,
	}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/api/v1/merchant/invoices/"+invoiceID, nil)
	r.SetPathValue("invoice_id", invoiceID)
	rec := callWithAPIKey(h.DB, h.GetMerchantInvoiceHandler, r, apiKey)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// ---------- DeleteMerchantInvoiceHandler ----------

func TestDeleteMerchantInvoiceHandler_NotFound(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)

	r := httptest.NewRequest(http.MethodDelete, "/api/v1/merchant/invoices/missing", nil)
	r.SetPathValue("invoice_id", "missing")
	rec := callWithAPIKey(h.DB, h.DeleteMerchantInvoiceHandler, r, apiKey)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteMerchantInvoiceHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	merchant := seedMerchant(t, h.DB)
	apiKey := seedMerchantAPIKey(t, h.DB, merchant.MerchantID)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)
	invoiceID := uuid.New().String()
	if err := h.DB.Create(&models.Invoice{
		InvoiceID:            invoiceID,
		InvoiceAmountCharged: decimal.NewFromInt(10),
		InvoiceStatus:        "created",
		InvoiceFeeAmount:     decimal.Zero,
		InvoiceFeeStatus:     "unpaid",
		InvoiceCryptoType:    "XRP",
		InvoiceCustomerID:    customer.CustomerID,
	}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	r := httptest.NewRequest(http.MethodDelete, "/api/v1/merchant/invoices/"+invoiceID, nil)
	r.SetPathValue("invoice_id", invoiceID)
	rec := callWithAPIKey(h.DB, h.DeleteMerchantInvoiceHandler, r, apiKey)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

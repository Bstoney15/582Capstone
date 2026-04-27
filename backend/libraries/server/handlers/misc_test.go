package routes

// Tests for the dashboard, search, health, widget, and CORS preflight handlers.

import (
	"backend/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------- HealthCheckHandler ----------

func TestHealthCheckHandler(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := callDirect(h.HealthCheckHandler, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "OK" {
		t.Errorf("body = %q, want OK", rec.Body.String())
	}
}

// ---------- GetCheckoutWidget ----------

func TestGetCheckoutWidget(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/widget/checkout.js", nil)
	rec := callDirect(h.GetCheckoutWidget, r)
	// Either 200 (asset present) or 404 (asset missing) is a valid handler response;
	// we just want to ensure the handler exits without panicking and sets a sane code.
	if rec.Code != http.StatusOK && rec.Code != http.StatusNotFound {
		t.Fatalf("unexpected code %d", rec.Code)
	}
	if rec.Code == http.StatusOK {
		ct := rec.Header().Get("Content-Type")
		if ct == "" {
			t.Errorf("expected Content-Type header to be set")
		}
	}
}

// ---------- WidgetPreflightHandler ----------

func TestWidgetPreflightHandler_NoOrigin(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodOptions, "/api/invoices", nil)
	rec := callDirect(h.WidgetPreflightHandler, r)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestWidgetPreflightHandler_WithOrigin(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodOptions, "/api/invoices", nil)
	r.Header.Set("Origin", "https://example.com")
	rec := callDirect(h.WidgetPreflightHandler, r)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected origin reflected, got %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

// ---------- GetDashboardHandler ----------

func TestGetDashboardHandler_Unauthorized(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	rec := callDirect(h.GetDashboardHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestGetDashboardHandler_MissingMerchantID(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	r := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetDashboardHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetDashboardHandler_Forbidden(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	r := httptest.NewRequest(http.MethodGet, "/api/dashboard?merchant_id=missing", nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetDashboardHandler, r)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestGetDashboardHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)

	r := httptest.NewRequest(http.MethodGet, "/api/dashboard?merchant_id="+merchant.MerchantID, nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.GetDashboardHandler, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ---------- SearchInvoiceHandler ----------

func TestSearchInvoiceHandler_Unauthorized(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/dashboard/search", nil)
	rec := callDirect(h.SearchInvoiceHandler, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestSearchInvoiceHandler_MissingMerchantID(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	r := httptest.NewRequest(http.MethodGet, "/api/dashboard/search", nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.SearchInvoiceHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestSearchInvoiceHandler_MissingInvoiceID(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	r := httptest.NewRequest(http.MethodGet, "/api/dashboard/search?merchant_id=mid", nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.SearchInvoiceHandler, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestSearchInvoiceHandler_Forbidden(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	r := httptest.NewRequest(http.MethodGet, "/api/dashboard/search?merchant_id=mid&invoice_id=abc", nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.SearchInvoiceHandler, r)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestSearchInvoiceHandler_Success(t *testing.T) {
	h := newTestHandler(t)
	user := seedUser(t, h.DB, "u", "pw")
	merchant := seedMerchant(t, h.DB)
	seedRole(t, h.DB, user.UserID, merchant.MerchantID, models.RoleAdmin)
	customer := seedCustomer(t, h.DB, merchant.MerchantID)
	mustSeedInvoice(t, h, customer.CustomerID, "paid")

	r := httptest.NewRequest(http.MethodGet, "/api/dashboard/search?merchant_id="+merchant.MerchantID+"&invoice_id=a", nil)
	r.AddCookie(loginAs(user.UserID))
	rec := callDirect(h.SearchInvoiceHandler, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

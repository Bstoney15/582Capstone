package routes

// Author: Charley Findling
// Created: 2026-03-30
// Description: Endpoint to fetch all invoices for a specific customer.

import (
	"backend/libraries/sessionManager"
	"backend/models"
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

// CustomerInvoiceResponse represents a single invoice in the billing history.
type CustomerInvoiceResponse struct {
	InvoiceID            string `json:"invoice_id"`
	InvoiceAmountCharged string `json:"invoice_amount_charged"`
	InvoiceStatus        string `json:"invoice_status"`
	InvoiceFeeAmount     string `json:"invoice_fee_amount"`
	InvoiceFeeStatus     string `json:"invoice_fee_status"`
	InvoiceCryptoType    string `json:"invoice_crypto_type"`
	InvoiceDateTime      string `json:"invoice_date_time"`
}

// GetCustomerInvoicesHandler retrieves all invoices linked to a specific customer.
// It verifies the requesting user has a role on the customer's merchant before returning data.
func (h *Handler) GetCustomerInvoicesHandler(w http.ResponseWriter, r *http.Request) {
	// ── JWT session authentication ──────────────────────
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

	// ── Extract customer ID from path ──────────────────
	customerID := r.PathValue("customer_id")
	if customerID == "" {
		http.Error(w, "customer_id is required", http.StatusBadRequest)
		return
	}

	// ── Look up the customer to find their merchant ────────────
	var customer models.Customer
	if err := h.DB.Where("customer_id = ?", customerID).First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Customer not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to look up customer", http.StatusInternalServerError)
		return
	}

	// ── Verify the user has ANY role on the merchant ───
	var role models.Role
	if err := h.DB.Where(
		"role_merchant_id = ? AND role_user_id = ?",
		customer.CustomerMerchantID,
		sessionData.UserID,
	).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Forbidden: you do not have access to this merchant", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to verify authorization", http.StatusInternalServerError)
		return
	}

	var invoices []models.Invoice
	if err := h.DB.Where("invoice_customer_id = ?", customerID).
		Order("invoice_date_time DESC").
		Find(&invoices).Error; err != nil {
		http.Error(w, "Failed to fetch invoices", http.StatusInternalServerError)
		return
	}

	// ── Build response ─────────────────────────────────────────
	responses := make([]CustomerInvoiceResponse, 0, len(invoices))
	for _, inv := range invoices {
		responses = append(responses, CustomerInvoiceResponse{
			InvoiceID:            inv.InvoiceID,
			InvoiceAmountCharged: inv.InvoiceAmountCharged.StringFixed(4),
			InvoiceStatus:        inv.InvoiceStatus,
			InvoiceFeeAmount:     inv.InvoiceFeeAmount.StringFixed(4),
			InvoiceFeeStatus:     inv.InvoiceFeeStatus,
			InvoiceCryptoType:    inv.InvoiceCryptoType,
			InvoiceDateTime:      inv.InvoiceDateTime.Format("2006-01-02T15:04:05Z"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}
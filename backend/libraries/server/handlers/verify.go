// verify.go – handler that accepts a transaction hash and queues background verification of an invoice payment.
package routes

// Author: Benjamin Stonestreet
// Created: 2026-03-01


import (
	"backend/models"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"gorm.io/gorm"
)

var xrplTxHashRegex = regexp.MustCompile(`^[A-Fa-f0-9]{64}$`)
var crossmarkPayloadIDRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// VerifyInvoiceRequest is the JSON payload for verifying an invoice payment.
type VerifyInvoiceRequest struct {
	InvoiceID string `json:"invoice_id"`
	TxHash    string `json:"tx_hash"`
}

// VerifyInvoicePaymentHandler queues an invoice verification in the background.
func (h *Handler) VerifyInvoicePaymentHandler(w http.ResponseWriter, r *http.Request) {
	applyWidgetCORSHeaders(w, r)

	var request VerifyInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if request.InvoiceID == "" || request.TxHash == "" {
		http.Error(w, "invoice_id and tx_hash are required", http.StatusBadRequest)
		return
	}

	var invoice models.Invoice
	if err := h.DB.Where("invoice_id = ?", request.InvoiceID).First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "invoice not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to fetch invoice", http.StatusInternalServerError)
		return
	}

	go h.verifyInvoiceInBackground(request.InvoiceID, request.TxHash)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "verification queued",
	})
}

// verifyInvoiceInBackground updates the invoice status based on the transaction hash format.
func (h *Handler) verifyInvoiceInBackground(invoiceID string, txHash string) {
	if xrplTxHashRegex.MatchString(txHash) {
		h.DB.Model(&models.Invoice{}).
			Where("invoice_id = ?", invoiceID).
			Update("invoice_status", "paid")
		return
	}

	if crossmarkPayloadIDRegex.MatchString(txHash) {
		h.DB.Model(&models.Invoice{}).
			Where("invoice_id = ?", invoiceID).
			Update("invoice_status", "verification_pending")
		return
	}

	h.DB.Model(&models.Invoice{}).
		Where("invoice_id = ?", invoiceID).
		Update("invoice_status", "verification_failed")
}

package routes

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

type VerifyInvoiceRequest struct {
	InvoiceID string `json:"invoice_id"`
	TxHash    string `json:"tx_hash"`
}

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

package routes

// Author: Benjamin Stonestreet
// Created: 2024-03-08

import (
	"backend/models"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// InvoiceStatusEvent represents the payload sent over SSE for an invoice update.
type InvoiceStatusEvent struct {
	InvoiceID string `json:"invoice_id"`
	Status    string `json:"status"`
}

// isTerminalInvoiceStatus returns true if the invoice status requires no further updates.
func isTerminalInvoiceStatus(status string) bool {
	return status == "paid" || status == "verification_failed"
}

// getInvoiceStatus retrieves the current status of an invoice from the database.
func (h *Handler) getInvoiceStatus(invoiceID string) (string, error) {
	var invoice models.Invoice
	err := h.DB.Select("invoice_status").Where("invoice_id = ?", invoiceID).First(&invoice).Error
	if err != nil {
		return "", err
	}

	return invoice.InvoiceStatus, nil
}

// writeSSEEvent serializes an InvoiceStatusEvent and flushes it to the client.
func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, payload InvoiceStatusEvent) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(body)); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

// StreamInvoiceEventsHandler opens a Server-Sent Events stream to push invoice updates to the client.
func (h *Handler) StreamInvoiceEventsHandler(w http.ResponseWriter, r *http.Request) {
	applyWidgetCORSHeaders(w, r)

	invoiceID := r.PathValue("uuid")
	if invoiceID == "" {
		http.Error(w, "invoice id is required", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	status, err := h.getInvoiceStatus(invoiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "invoice not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to resolve invoice", http.StatusInternalServerError)
		return
	}

	lastStatus := ""
	if err := writeSSEEvent(w, flusher, InvoiceStatusEvent{InvoiceID: invoiceID, Status: status}); err != nil {
		return
	}
	lastStatus = status

	if isTerminalInvoiceStatus(status) {
		return
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.NewTimer(4 * time.Minute)
	defer timeout.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-timeout.C:
			return
		case <-ticker.C:
			updatedStatus, statusErr := h.getInvoiceStatus(invoiceID)
			if statusErr != nil {
				return
			}

			if updatedStatus == lastStatus {
				continue
			}

			if err := writeSSEEvent(w, flusher, InvoiceStatusEvent{InvoiceID: invoiceID, Status: updatedStatus}); err != nil {
				return
			}

			lastStatus = updatedStatus
			if isTerminalInvoiceStatus(updatedStatus) {
				return
			}
		}
	}
}
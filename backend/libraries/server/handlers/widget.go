// widget.go – handler that serves the embedded checkout widget JavaScript file to clients.
package routes

// Author: Benjamin Stonestreet
// Created: 2026-03-01

import (
	widgetstatic "backend/static/widget"
	"net/http"
)

// GetCheckoutWidget serves the static MyPay checkout script payload to clients.
func (h *Handler) GetCheckoutWidget(w http.ResponseWriter, r *http.Request) {
	if len(widgetstatic.CheckoutJS) == 0 {
		http.Error(w, "checkout widget not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(widgetstatic.CheckoutJS)
}

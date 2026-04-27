// cors.go – applies CORS headers for the checkout widget and handles OPTIONS preflight requests.
package routes

// Author: Benjamin Stonestreet
// Created: 2026-03-01

import "net/http"

// applyWidgetCORSHeaders sets the CORS response headers needed by the embedded
// checkout widget. It reflects the request Origin rather than using "*" so that
// credentialed requests are accepted by browsers.
func applyWidgetCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Vary", "Origin")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, X-API-Key")
	w.Header().Set("Access-Control-Max-Age", "600")
}

// WidgetPreflightHandler responds to CORS OPTIONS preflight requests from the checkout widget.
func (h *Handler) WidgetPreflightHandler(w http.ResponseWriter, r *http.Request) {
	applyWidgetCORSHeaders(w, r)
	w.WriteHeader(http.StatusNoContent)
}

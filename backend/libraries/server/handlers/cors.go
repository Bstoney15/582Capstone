package routes

// author: Benjamin Stonestreet
// created: 2024-03-08
// description:
// This file implements the WidgetPreflightHandler, which handles CORS preflight
// requests from the XRPay checkout widget. It applies the necessary CORS headers
// to allow cross-origin requests from the widget's domain and responds with a
// 204 No Content status.


import "net/http"

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

func (h *Handler) WidgetPreflightHandler(w http.ResponseWriter, r *http.Request) {
	applyWidgetCORSHeaders(w, r)
	w.WriteHeader(http.StatusNoContent)
}

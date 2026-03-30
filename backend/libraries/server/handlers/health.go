package routes

// Author: Benjamin Stonestreet
// Created: 2026-02-02
// Description:
// This file implements the HealthCheckHandler, which provides a simple endpoint
// for monitoring the health of the server. It responds with a 200 OK status and
// a plain text "OK" message when the server is running properly.

import (
	"net/http"
)

// HealthCheckHandler responds to health check pings.
func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
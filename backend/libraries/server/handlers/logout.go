// logout.go – handler that clears the caller's session cookie and ends their authenticated session.
package routes

// Author: Benjamin Stonestreet
// Created: 2026-04-07

import (
	"backend/libraries/sessionManager"
	"encoding/json"
	"net/http"
)

// LogoutHandler clears the caller's session cookie and returns a success payload.
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	sessionManager.ClearSessionCookie(w)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logout successful",
	})
}

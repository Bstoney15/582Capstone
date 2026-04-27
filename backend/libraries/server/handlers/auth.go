// auth.go – handler that verifies the caller's session cookie and returns authentication status.
package routes

// Author: Benjamin Stonestreet
// Created: 2026-02-20
// Description:
// This file implements the CheckAuthHandler, which verifies the user's session
// and returns their authentication status. It checks for a valid session token
// in the request cookies, validates it, and responds with a JSON object
// indicating whether the user is authenticated and their user ID if so.

import (
	"backend/libraries/sessionManager"
	"encoding/json"
	"net/http"
)

// AuthResponse is the JSON body returned by CheckAuthHandler.
// UserID is omitted from the response when the user is not authenticated.
type AuthResponse struct {
	Authenticated bool   `json:"authenticated"`
	UserID        string `json:"userId,omitempty"`
}

// CheckAuthHandler verifies the caller's session and returns their
// authentication status. A 401 is returned when no session cookie is present
// or the session has expired. On success a 200 is returned with the user's ID.
func (h *Handler) CheckAuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract the session token from the request cookie.
	sessionToken, err := sessionManager.GetSessionToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthResponse{Authenticated: false})
		return
	}

	// Validate the session; CheckSession also handles sliding-window rotation.
	sessionData, rotatedToken, active := sessionManager.CheckSession(sessionToken)
	if !active {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthResponse{Authenticated: false})
		return
	}

	// If the session was rotated, push the new token to the client.
	if rotatedToken != "" {
		sessionManager.SetSessionCookie(w, rotatedToken)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{
		Authenticated: true,
		UserID:        sessionData.UserID,
	})
}

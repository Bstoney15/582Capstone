// File contains functions for managing session instances using JWT
//
// Author: Benjamin Stonestreet
// Date: 2025-10-23
// Refactored to JWT: 2026-02-20

package sessionManager

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const sessionLength = 600 // session length in seconds (10 minutes)
const SessionCookieName = "session_id"

// secret key used to sign the tokens. In a production app, this should be an environment variable.
var jwtKey = []byte(getJWTKey())

// getJWTKey retrieves the JWT secret key from the environment,
// or falls back to a default value if not set.
func getJWTKey() string {
	key := os.Getenv("JWT_SECRET_KEY")
	if key == "" {
		// Provide a fallback for local development if the environment variable map isn't set yet
		return "my_secret_key"
	}
	return key
}

// sessionData stores parsed token information including the user ID,
// the full token string, and its expiration time.
type sessionData struct {
	UserID    string
	SessionID string
	Expiry    time.Time
}

// SessionClaims represents the structure of the JWT payload
type SessionClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// creates a new session for the given user ID and returns the JWT token
func CreateSession(userID string) string {
	expirationTime := time.Now().Add(sessionLength * time.Second)

	claims := &SessionClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		fmt.Println("Error creating token:", err)
		return ""
	}

	return tokenString
}

// retrieves session data for the given JWT token.
// Also returns a string containing a new token if the current one was close to expiring and got rotated.
func CheckSession(sessionToken string) (sessionData, string, bool) {
	claims := &SessionClaims{}

	token, err := jwt.ParseWithClaims(sessionToken, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return sessionData{}, "", false
	}

	expiry := claims.ExpiresAt.Time
	var newToken string

	// Rotate the token if it has less than a minute until expiry
	if time.Until(expiry) < 1*time.Minute {
		newToken = CreateSession(claims.UserID)
		// Update the session details since we are issuing a new token
		sessionToken = newToken
		expiry = time.Now().Add(sessionLength * time.Second)
	}

	return sessionData{
		UserID:    claims.UserID,
		SessionID: sessionToken,
		Expiry:    expiry,
	}, newToken, true
}

// SetSessionCookie sets the JWT token in a secure http-only cookie
func SetSessionCookie(w http.ResponseWriter, sessionToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionToken,
		HttpOnly: true,
		Path:     "/",
		// Secure:   true, // Uncomment when running on HTTPS
		// SameSite: http.SameSiteStrictMode,
		MaxAge: sessionLength,
	})
}

// GetSessionToken retrieves the session token from the request cookies
func GetSessionToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

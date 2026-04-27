// sessionManager.go – JWT-based session management: create, validate, rotate, and clear session cookies.
package sessionManager

// Author: Benjamin Stonestreet
// Created: 2026-02-02

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// sessionLength is the session duration in seconds (10 minutes).
const sessionLength = 600

// SessionCookieName is the name of the HTTP cookie that carries the JWT session token.
const SessionCookieName = "session_id"

// jwtKey is the HMAC signing key, loaded from the JWT_SECRET_KEY environment variable.
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

// SessionClaims represents the structure of the JWT payload.
type SessionClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// CreateSession creates a new session for the given user ID and returns the signed JWT token.
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

// CheckSession validates the given JWT token and returns the session data.
// If the token is close to expiring it is automatically rotated; the second
// return value is the new token string (empty if no rotation occurred).
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

	// Rotate the token if it has less than a minute until expiry.
	if time.Until(expiry) < 1*time.Minute {
		newToken = CreateSession(claims.UserID)
		// Update the session details since we are issuing a new token.
		sessionToken = newToken
		expiry = time.Now().Add(sessionLength * time.Second)
	}

	return sessionData{
		UserID:    claims.UserID,
		SessionID: sessionToken,
		Expiry:    expiry,
	}, newToken, true
}

// SetSessionCookie sets the JWT token in a secure http-only cookie.
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

// ClearSessionCookie expires the session cookie so the browser removes it.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		// Secure:   true, // Uncomment when running on HTTPS
		// SameSite: http.SameSiteStrictMode,
		MaxAge:  -1,
		Expires: time.Unix(0, 0),
	})
}

// GetSessionToken retrieves the session token from the request cookies.
func GetSessionToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

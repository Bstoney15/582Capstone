// merchant_api_key.go – middleware that validates merchant API keys and injects the merchant ID into the request context.
package apiauth

// Author: Benjamin Stonestreet
// Created: 2026-04-12

import (
	"backend/models"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

// maxRequestBodyBytes caps the number of bytes read from a request body when
// searching for an embedded api_key field, preventing excessively large reads.
const maxRequestBodyBytes = 1024 * 1024

// contextKey is a private type used for context value keys to avoid collisions.
type contextKey string

// merchantIDContextKey is the key under which the validated merchant ID is stored in the request context.
const merchantIDContextKey contextKey = "merchant_api_key_merchant_id"

// MerchantIDFromContext returns the merchant id attached by API key middleware.
func MerchantIDFromContext(ctx context.Context) (string, bool) {
	merchantID, ok := ctx.Value(merchantIDContextKey).(string)
	if !ok || strings.TrimSpace(merchantID) == "" {
		return "", false
	}

	return merchantID, true
}

// RequireMerchantAPIKey validates merchant API keys and blocks unauthorized requests.
func RequireMerchantAPIKey(db *gorm.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := extractMerchantAPIKey(r)
		if err != nil || strings.TrimSpace(apiKey) == "" {
			http.Error(w, "invalid merchant api key", http.StatusUnauthorized)
			return
		}

		hashed := hashAPIKey(apiKey)

		var record models.MerchantAPIKey
		queryErr := db.Where(
			"merchant_api_key_hashed IN ? AND merchant_api_key_revoked = ?",
			[]string{hashed, apiKey},
			false,
		).First(&record).Error
		if queryErr != nil {
			if errors.Is(queryErr, gorm.ErrRecordNotFound) {
				http.Error(w, "invalid merchant api key", http.StatusUnauthorized)
				return
			}

			http.Error(w, "failed to validate merchant api key", http.StatusInternalServerError)
			return
		}

		merchantID := strings.TrimSpace(record.MerchantAPIKeyMerchantID)
		if merchantID == "" {
			http.Error(w, "invalid merchant api key", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), merchantIDContextKey, merchantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractMerchantAPIKey searches for an API key in the request, checking the
// X-API-Key header, Authorization bearer token, api_key query param, and JSON body in that order.
func extractMerchantAPIKey(r *http.Request) (string, error) {
	apiKey := strings.TrimSpace(r.Header.Get("X-API-Key"))
	if apiKey != "" {
		return apiKey, nil
	}

	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
		token := strings.TrimSpace(authorization[7:])
		if token != "" {
			return token, nil
		}
	}

	apiKey = strings.TrimSpace(r.URL.Query().Get("api_key"))
	if apiKey != "" {
		return apiKey, nil
	}

	if r.Body == nil {
		return "", nil
	}

	contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if !strings.Contains(contentType, "application/json") {
		return "", nil
	}

	// Read and buffer the body so downstream handlers can still read it.
	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBodyBytes))
	if err != nil {
		return "", err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	if len(body) == 0 {
		return "", nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", nil
	}

	raw, ok := payload["merchant_api_key"]
	if !ok {
		return "", nil
	}

	asString, ok := raw.(string)
	if !ok {
		return "", nil
	}

	return strings.TrimSpace(asString), nil
}

// hashAPIKey returns the SHA-256 hex digest of the given plaintext token.
func hashAPIKey(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

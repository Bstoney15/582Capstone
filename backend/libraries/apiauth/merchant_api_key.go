// Authors: Ben Stonestreet
// Created: 04/12/26
// Description: works with merchant api keys
package apiauth

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

const maxRequestBodyBytes = 1024 * 1024

type contextKey string

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

func hashAPIKey(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

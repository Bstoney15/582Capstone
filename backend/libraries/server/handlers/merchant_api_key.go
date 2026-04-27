// Authors: Ben Stonestreet
// Created: 03/29/26
// Description: handler functions for merchant api keys
package routes

import (
	"backend/libraries/sessionManager"
	"backend/models"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type merchantAPIKeySummary struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	GeneratedAt time.Time `json:"generated_at"`
}

type createMerchantAPIKeyRequest struct {
	MerchantID string `json:"merchant_id"`
	Name       string `json:"name"`
}

type createMerchantAPIKeyResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	GeneratedAt time.Time `json:"generated_at"`
	APIKey      string    `json:"api_key"`
}

func hashAPIKey(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

func generatePlaintextAPIKey() (string, error) {
	randomBytes := make([]byte, 24)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return "sk_" + hex.EncodeToString(randomBytes), nil
}

func (h *Handler) requireMerchantMembership(r *http.Request, merchantID string) (string, bool) {
	sessionToken, err := sessionManager.GetSessionToken(r)
	if err != nil {
		return "", false
	}

	sessionData, _, active := sessionManager.CheckSession(sessionToken)
	if !active {
		return "", false
	}

	var role models.Role
	if err := h.DB.Where("role_merchant_id = ? AND role_user_id = ?", merchantID, sessionData.UserID).First(&role).Error; err != nil {
		return "", false
	}

	return sessionData.UserID, true
}

func (h *Handler) GetMerchantAPIKeysHandler(w http.ResponseWriter, r *http.Request) {
	merchantID := strings.TrimSpace(r.URL.Query().Get("merchant_id"))
	if merchantID == "" {
		http.Error(w, "merchant_id is required", http.StatusBadRequest)
		return
	}

	if _, ok := h.requireMerchantMembership(r, merchantID); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var keys []models.MerchantAPIKey
	if err := h.DB.Where("merchant_api_key_merchant_id = ? AND merchant_api_key_revoked = ?", merchantID, false).Order("merchant_api_key_created_at DESC").Find(&keys).Error; err != nil {
		http.Error(w, "Failed to fetch API keys", http.StatusInternalServerError)
		return
	}

	response := make([]merchantAPIKeySummary, 0, len(keys))
	for _, key := range keys {
		response = append(response, merchantAPIKeySummary{
			ID:          key.MerchantAPIKeyID,
			Name:        key.MerchantAPIKeyName,
			GeneratedAt: key.MerchantAPIKeyCreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) CreateMerchantAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody createMerchantAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}

	requestBody.MerchantID = strings.TrimSpace(requestBody.MerchantID)
	requestBody.Name = strings.TrimSpace(requestBody.Name)
	if requestBody.MerchantID == "" || requestBody.Name == "" {
		http.Error(w, "merchant_id and name are required", http.StatusBadRequest)
		return
	}

	if _, ok := h.requireMerchantMembership(r, requestBody.MerchantID); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	plainAPIKey, err := generatePlaintextAPIKey()
	if err != nil {
		http.Error(w, "Failed to generate API key", http.StatusInternalServerError)
		return
	}

	dbKey := models.MerchantAPIKey{
		MerchantAPIKeyID:         uuid.New().String(),
		MerchantAPIKeyName:       requestBody.Name,
		MerchantAPIKeyHashed:     hashAPIKey(plainAPIKey),
		MerchantAPIKeyRevoked:    false,
		MerchantAPIKeyMerchantID: requestBody.MerchantID,
	}

	if err := h.DB.Create(&dbKey).Error; err != nil {
		http.Error(w, "Failed to create API key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createMerchantAPIKeyResponse{
		ID:          dbKey.MerchantAPIKeyID,
		Name:        dbKey.MerchantAPIKeyName,
		GeneratedAt: dbKey.MerchantAPIKeyCreatedAt,
		APIKey:      plainAPIKey,
	})
}

func (h *Handler) DeleteMerchantAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	apiKeyID := strings.TrimSpace(r.PathValue("api_key"))
	merchantID := strings.TrimSpace(r.URL.Query().Get("merchant_id"))
	if apiKeyID == "" {
		http.Error(w, "api_key is required", http.StatusBadRequest)
		return
	}
	if merchantID == "" {
		http.Error(w, "merchant_id is required", http.StatusBadRequest)
		return
	}

	if _, ok := h.requireMerchantMembership(r, merchantID); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var existing models.MerchantAPIKey
	if err := h.DB.Where("merchant_api_key_id = ? AND merchant_api_key_merchant_id = ? AND merchant_api_key_revoked = ?", apiKeyID, merchantID, false).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "API key not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch API key", http.StatusInternalServerError)
		return
	}

	if err := h.DB.Model(&existing).Update("merchant_api_key_revoked", true).Error; err != nil {
		http.Error(w, "Failed to delete API key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

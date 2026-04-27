package routes

import (
	"backend/models"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type webhookSummary struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	HasSecret bool   `json:"hasSecret"`
	UpdatedAt string `json:"updatedAt"`
}

func (h *Handler) GetMerchantWebhooksHandler(w http.ResponseWriter, r *http.Request) {
	merchantID := strings.TrimSpace(r.URL.Query().Get("merchant_id"))
	if merchantID == "" {
		http.Error(w, "merchant_id is required", http.StatusBadRequest)
		return
	}

	if _, ok := h.requireMerchantMembership(r, merchantID); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var config models.MerchantWebhookKey
	err := h.DB.Where("merchant_webhook_key_merchant_id = ?", merchantID).Limit(1).Find(&config).Error
	if err != nil {
		http.Error(w, "failed to fetch webhook", http.StatusInternalServerError)
		return
	}

	result := []webhookSummary{}
	if config.MerchantWebhookKeyID != "" && strings.TrimSpace(config.MerchantWebhookURL) != "" {
		result = append(result, webhookSummary{
			ID:        config.MerchantWebhookKeyID,
			URL:       config.MerchantWebhookURL,
			HasSecret: strings.TrimSpace(config.MerchantWebhookKey) != "",
			UpdatedAt: config.MerchantWebhookKeyID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) CreateMerchantWebhookHandler(w http.ResponseWriter, r *http.Request) {
	type createWebhookRequest struct {
		MerchantID string `json:"merchant_id"`
		URL        string `json:"url"`
		Secret     string `json:"secret"`
	}

	var body createWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	body.MerchantID = strings.TrimSpace(body.MerchantID)
	body.URL = strings.TrimSpace(body.URL)
	body.Secret = strings.TrimSpace(body.Secret)

	if body.MerchantID == "" || body.URL == "" || body.Secret == "" {
		http.Error(w, "merchant_id, url, and secret are required", http.StatusBadRequest)
		return
	}

	if _, ok := h.requireMerchantMembership(r, body.MerchantID); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var existing models.MerchantWebhookKey
	err := h.DB.Where("merchant_webhook_key_merchant_id = ?", body.MerchantID).Limit(1).Find(&existing).Error
	if err != nil {
		http.Error(w, "failed to check existing webhook", http.StatusInternalServerError)
		return
	}

	if existing.MerchantWebhookKeyID != "" {
		existing.MerchantWebhookURL = body.URL
		existing.MerchantWebhookKey = body.Secret
		if err := h.DB.Save(&existing).Error; err != nil {
			http.Error(w, "failed to update webhook", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(webhookSummary{
			ID:        existing.MerchantWebhookKeyID,
			URL:       existing.MerchantWebhookURL,
			HasSecret: true,
			UpdatedAt: existing.MerchantWebhookKeyID,
		})
		return
	}

	record := models.MerchantWebhookKey{
		MerchantWebhookKeyID:         uuid.New().String(),
		MerchantWebhookURL:           body.URL,
		MerchantWebhookKey:           body.Secret,
		MerchantWebhookKeyMerchantID: body.MerchantID,
	}

	if err := h.DB.Create(&record).Error; err != nil {
		http.Error(w, "failed to create webhook", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(webhookSummary{
		ID:        record.MerchantWebhookKeyID,
		URL:       record.MerchantWebhookURL,
		HasSecret: true,
		UpdatedAt: record.MerchantWebhookKeyID,
	})
}

func (h *Handler) DeleteMerchantWebhookHandler(w http.ResponseWriter, r *http.Request) {
	webhookID := strings.TrimSpace(r.PathValue("webhook_id"))
	merchantID := strings.TrimSpace(r.URL.Query().Get("merchant_id"))

	if webhookID == "" || merchantID == "" {
		http.Error(w, "webhook_id and merchant_id are required", http.StatusBadRequest)
		return
	}

	if _, ok := h.requireMerchantMembership(r, merchantID); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var record models.MerchantWebhookKey
	if err := h.DB.Where("merchant_webhook_key_id = ? AND merchant_webhook_key_merchant_id = ?", webhookID, merchantID).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "webhook not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to fetch webhook", http.StatusInternalServerError)
		return
	}

	if err := h.DB.Delete(&record).Error; err != nil {
		http.Error(w, "failed to delete webhook", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

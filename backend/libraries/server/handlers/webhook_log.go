package routes

import (
	"backend/libraries/webhooks"
	"backend/models"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type webhookLogSummary struct {
	ID         string    `json:"id"`
	InvoiceID  string    `json:"invoice_id"`
	EventType  string    `json:"event_type"`
	StatusCode int       `json:"status_code"`
	Attempts   int       `json:"attempts"`
	Succeeded  bool      `json:"succeeded"`
	Error      string    `json:"error"`
	CreatedAt  time.Time `json:"created_at"`
}

func (h *Handler) ListWebhookLogsHandler(w http.ResponseWriter, r *http.Request) {
	merchantID := strings.TrimSpace(r.URL.Query().Get("merchant_id"))
	if merchantID == "" {
		http.Error(w, "merchant_id is required", http.StatusBadRequest)
		return
	}

	if _, ok := h.requireMerchantMembership(r, merchantID); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var logs []models.WebhookLog
	if err := h.DB.Where("webhook_log_merchant_id = ?", merchantID).
		Order("webhook_log_created_at DESC").
		Limit(100).
		Find(&logs).Error; err != nil {
		http.Error(w, "failed to fetch webhook logs", http.StatusInternalServerError)
		return
	}

	response := make([]webhookLogSummary, 0, len(logs))
	for _, l := range logs {
		response = append(response, webhookLogSummary{
			ID:         l.WebhookLogID,
			InvoiceID:  l.WebhookLogInvoiceID,
			EventType:  l.WebhookLogEventType,
			StatusCode: l.WebhookLogStatusCode,
			Attempts:   l.WebhookLogAttempts,
			Succeeded:  l.WebhookLogSucceeded,
			Error:      l.WebhookLogError,
			CreatedAt:  l.WebhookLogCreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) ResendWebhookHandler(w http.ResponseWriter, r *http.Request) {
	logID := strings.TrimSpace(r.PathValue("log_id"))
	merchantID := strings.TrimSpace(r.URL.Query().Get("merchant_id"))

	if logID == "" || merchantID == "" {
		http.Error(w, "log_id and merchant_id are required", http.StatusBadRequest)
		return
	}

	if _, ok := h.requireMerchantMembership(r, merchantID); !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var original models.WebhookLog
	if err := h.DB.Where("webhook_log_id = ? AND webhook_log_merchant_id = ?", logID, merchantID).First(&original).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "webhook log not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to fetch webhook log", http.StatusInternalServerError)
		return
	}

	var config models.MerchantWebhookKey
	if err := h.DB.Where("merchant_webhook_key_merchant_id = ?", merchantID).Limit(1).Find(&config).Error; err != nil || strings.TrimSpace(config.MerchantWebhookURL) == "" {
		http.Error(w, "no webhook configured for this merchant", http.StatusBadRequest)
		return
	}

	var payload interface{}
	if err := json.Unmarshal([]byte(original.WebhookLogPayload), &payload); err != nil {
		http.Error(w, "failed to parse original payload", http.StatusInternalServerError)
		return
	}

	dispatcher := webhooks.NewDispatcher(webhooks.DispatcherConfig{})
	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()

	result, dispatchErr := dispatcher.Dispatch(ctx, config.MerchantWebhookURL, config.MerchantWebhookKey, original.WebhookLogEventType, payload)

	newLog := models.WebhookLog{
		WebhookLogID:         uuid.New().String(),
		WebhookLogMerchantID: merchantID,
		WebhookLogInvoiceID:  original.WebhookLogInvoiceID,
		WebhookLogEventType:  original.WebhookLogEventType,
		WebhookLogPayload:    original.WebhookLogPayload,
		WebhookLogSucceeded:  dispatchErr == nil,
	}
	if result != nil {
		newLog.WebhookLogStatusCode = result.StatusCode
		newLog.WebhookLogAttempts = result.Attempt
	}
	if dispatchErr != nil {
		newLog.WebhookLogError = dispatchErr.Error()
	}

	if err := h.DB.Create(&newLog).Error; err != nil {
		http.Error(w, "failed to save resend log", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if dispatchErr != nil {
		w.WriteHeader(http.StatusBadGateway)
	}
	json.NewEncoder(w).Encode(webhookLogSummary{
		ID:         newLog.WebhookLogID,
		InvoiceID:  newLog.WebhookLogInvoiceID,
		EventType:  newLog.WebhookLogEventType,
		StatusCode: newLog.WebhookLogStatusCode,
		Attempts:   newLog.WebhookLogAttempts,
		Succeeded:  newLog.WebhookLogSucceeded,
		Error:      newLog.WebhookLogError,
		CreatedAt:  newLog.WebhookLogCreatedAt,
	})
}

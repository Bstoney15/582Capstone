// webhook_log.go – GORM model recording each outbound webhook dispatch attempt and its result.
package models

// Author: Benjamin Stonestreet
// Created: 2026-04-26

import "time"

// WebhookLog records each outbound webhook dispatch attempt, including the HTTP status,
// number of retries, and any error message returned by the merchant endpoint.
type WebhookLog struct {
	WebhookLogID         string    `gorm:"primaryKey;type:varchar(36)" json:"webhook_log_id"`
	WebhookLogMerchantID string    `gorm:"not null;type:varchar(36);index" json:"webhook_log_merchant_id"`
	WebhookLogInvoiceID  string    `gorm:"type:varchar(36);default:''" json:"webhook_log_invoice_id"`
	WebhookLogEventType  string    `gorm:"not null;default:''" json:"webhook_log_event_type"`
	WebhookLogPayload    string    `gorm:"type:text;not null;default:''" json:"webhook_log_payload"`
	WebhookLogStatusCode int       `gorm:"default:0" json:"webhook_log_status_code"`
	WebhookLogAttempts   int       `gorm:"default:0" json:"webhook_log_attempts"`
	WebhookLogSucceeded  bool      `gorm:"not null;default:false" json:"webhook_log_succeeded"`
	WebhookLogError      string    `gorm:"type:text;default:''" json:"webhook_log_error"`
	WebhookLogCreatedAt  time.Time `gorm:"autoCreateTime" json:"webhook_log_created_at"`
	Merchant             Merchant  `gorm:"foreignKey:WebhookLogMerchantID;references:MerchantID" json:"merchant,omitempty"`
}

// TableName sets the table name for WebhookLog.
func (WebhookLog) TableName() string {
	return "webhook_log"
}

package models

// Author: Benjamin Stonestreet
// Created: 2024-02-02

// MerchantWebhookKey holds the webhook keys for a merchant, used to securely
// send event notifications to the merchant's endpoints.
type MerchantWebhookKey struct {
	MerchantWebhookKeyID         string   `gorm:"primaryKey;type:varchar(36)" json:"merchant_webhook_key_id"`
	MerchantWebhookKey           string   `gorm:"not null" json:"merchant_webhook_key"`
	MerchantWebhookKeyMerchantID string   `gorm:"not null;type:varchar(36)" json:"merchant_webhook_key_merchant_id"`
	Merchant                     Merchant `gorm:"foreignKey:MerchantWebhookKeyMerchantID;references:MerchantID" json:"merchant,omitempty"`
}

// TableName overrides the default GORM table name to "merchant_webhook_key".
func (MerchantWebhookKey) TableName() string {
	return "merchant_webhook_key"
}

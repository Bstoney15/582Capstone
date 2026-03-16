package models

// Author: Benjamin Stonestreet
// Created: 2024-02-02

// MerchantAPIKey stores a hashed API key that allows merchants to authenticate
// requests to the MyPay API.
type MerchantAPIKey struct {
	MerchantAPIKeyID         string   `gorm:"primaryKey;type:varchar(36)" json:"merchant_api_key_id"`
	MerchantAPIKeyHashed     string   `gorm:"not null" json:"merchant_api_key_hashed"`
	MerchantAPIKeyMerchantID string   `gorm:"not null;type:varchar(36)" json:"merchant_api_key_merchant_id"`
	Merchant                 Merchant `gorm:"foreignKey:MerchantAPIKeyMerchantID;references:MerchantID" json:"merchant,omitempty"`
}

// TableName sets the table name for MerchantAPIKey.
func (MerchantAPIKey) TableName() string {
	return "merchant_api_key"
}

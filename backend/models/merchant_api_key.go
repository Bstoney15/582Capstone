package models

type MerchantAPIKey struct {
	MerchantAPIKeyID         string   `gorm:"primaryKey" json:"merchant_api_key_id"`
	MerchantAPIKeyHashed     string   `gorm:"not null" json:"merchant_api_key_hashed"`
	MerchantAPIKeyMerchantID string   `gorm:"not null" json:"merchant_api_key_merchant_id"`
	Merchant                 Merchant `gorm:"foreignKey:MerchantAPIKeyMerchantID;references:MerchantID" json:"merchant,omitempty"`
}

func (MerchantAPIKey) TableName() string {
	return "merchant_api_key"
}

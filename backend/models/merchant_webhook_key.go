package models

type MerchantWebhookKey struct {
	MerchantWebhookKeyID         string   `gorm:"primaryKey;type:varchar(36)" json:"merchant_webhook_key_id"`
	MerchantWebhookKey           string   `gorm:"not null" json:"merchant_webhook_key"`
	MerchantWebhookKeyMerchantID string   `gorm:"not null;type:varchar(36)" json:"merchant_webhook_key_merchant_id"`
	Merchant                     Merchant `gorm:"foreignKey:MerchantWebhookKeyMerchantID;references:MerchantID" json:"merchant,omitempty"`
}

func (MerchantWebhookKey) TableName() string {
	return "merchant_webhook_key"
}

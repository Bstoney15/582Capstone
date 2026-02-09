package models

type Merchant struct {
	MerchantID   string `gorm:"primaryKey" json:"merchant_id"`
	MerchantName string `gorm:"not null" json:"merchant_name"`
}

func (Merchant) TableName() string {
	return "merchant"
}

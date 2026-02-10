package models

type MerchantAddress struct {
	MerchantAddressID         string   `gorm:"primaryKey;type:varchar(36)" json:"merchant_address_id"`
	MerchantAddressMerchantID string   `gorm:"not null;type:varchar(36)" json:"merchant_address_merchant_id"`
	Merchant                  Merchant `gorm:"foreignKey:MerchantAddressMerchantID;references:MerchantID" json:"merchant,omitempty"`
	MerchantAddressLine1      string   `gorm:"not null" json:"merchant_address_line_1"`
	MerchantAddressLine2      string   `gorm:"not null" json:"merchant_address_line_2"`
	MerchantAddressCity       string   `gorm:"not null" json:"merchant_address_city"`
	MerchantAddressState      string   `gorm:"not null" json:"merchant_address_state"`
	MerchantAddressPostalCode int      `gorm:"not null" json:"merchant_address_postal_code"`
	MerchantAddressVerified   bool     `gorm:"not null" json:"merchant_address_verified"`
}

func (MerchantAddress) TableName() string {
	return "merchant_address"
}

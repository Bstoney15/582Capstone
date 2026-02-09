package models

import "time"

type MerchantOwner struct {
	MerchantOwnerID          string    `gorm:"primaryKey" json:"merchant_owner_id"`
	MerchantOwnerMerchantID  string    `gorm:"not null" json:"merchant_owner_merchant_id"`
	Merchant                 Merchant  `gorm:"foreignKey:MerchantOwnerMerchantID;references:MerchantID" json:"merchant,omitempty"`
	MerchantOwnerFirstName   string    `gorm:"not null" json:"merchant_owner_first_name"`
	MerchantOwnerLastName    string    `gorm:"not null" json:"merchant_owner_last_name"`
	MerchantOwnerPhoneNumber string    `gorm:"not null" json:"merchant_owner_phone_number"`
	MerchantOwnerEmail       string    `gorm:"not null" json:"merchant_owner_email"`
	MerchantOwnerDOB         time.Time `gorm:"not null" json:"merchant_owner_dob"`
	MerchantOwnerStake       float64   `gorm:"not null" json:"merchant_owner_stake"`
}

func (MerchantOwner) TableName() string {
	return "merchant_owner"
}

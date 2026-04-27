// merchant_owner.go – GORM model storing information about individual owners or stakeholders of a merchant.
package models

// Author: Benjamin Stonestreet
// Created: 2026-02-09

import (
	"time"

	"github.com/shopspring/decimal"
)

// MerchantOwner stores information about an individual owner or stakeholder
// associated with a specific merchant.
type MerchantOwner struct {
	MerchantOwnerID          string          `gorm:"primaryKey;type:varchar(36)" json:"merchant_owner_id"`
	MerchantOwnerMerchantID  string          `gorm:"not null;type:varchar(36)" json:"merchant_owner_merchant_id"`
	Merchant                 Merchant        `gorm:"foreignKey:MerchantOwnerMerchantID;references:MerchantID" json:"merchant,omitempty"`
	MerchantOwnerFirstName   string          `gorm:"not null" json:"merchant_owner_first_name"`
	MerchantOwnerLastName    string          `gorm:"not null" json:"merchant_owner_last_name"`
	MerchantOwnerPhoneNumber string          `gorm:"not null" json:"merchant_owner_phone_number"`
	MerchantOwnerEmail       string          `gorm:"not null" json:"merchant_owner_email"`
	MerchantOwnerDOB         time.Time       `gorm:"not null" json:"merchant_owner_dob"`
	MerchantOwnerStake       decimal.Decimal `gorm:"not null;type:decimal(16,4)" json:"merchant_owner_stake"`
}

// TableName overrides the default GORM table name to "merchant_owner".
func (MerchantOwner) TableName() string {
	return "merchant_owner"
}

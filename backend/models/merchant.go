package models

// Author: Benjamin Stonestreet
// Created: 2026-02-02

// Merchant stores the core identifier and name of a registered merchant.
type Merchant struct {
	MerchantID   string `gorm:"primaryKey;type:varchar(36)" json:"merchant_id"`
	MerchantName string `gorm:"not null" json:"merchant_name"`
}

// TableName sets the table name for Merchant.
func (Merchant) TableName() string {
	return "merchant"
}

package models

// Author: Benjamin Stonestreet
// Created: 2024-02-02

import "github.com/shopspring/decimal"

// Deposit tracks an incoming transaction from a customer meant to be added to their balance.
type Deposit struct {
	DepositID         string          `gorm:"primaryKey;type:varchar(36)" json:"deposit_id"`
	DepositAmount     decimal.Decimal `gorm:"not null;type:decimal(16,4)" json:"deposit_amount"`
	DepositStatus     string          `gorm:"not null" json:"deposit_status"`
	DepositFeeAmount  decimal.Decimal `gorm:"not null;type:decimal(16,4)" json:"deposit_fee_amount"`
	DepositFeeStatus  string          `gorm:"not null" json:"deposit_fee_status"`
	DepositCryptoType string          `gorm:"not null" json:"deposit_crypto_type"`
	DepositCustomerID string          `gorm:"not null;type:varchar(36)" json:"deposit_customer_id"`
	Customer          Customer        `gorm:"foreignKey:DepositCustomerID;references:CustomerID" json:"customer,omitempty"`
}

// TableName sets the table name for Deposit.
func (Deposit) TableName() string {
	return "deposit"
}

package models

import "github.com/shopspring/decimal"

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

func (Deposit) TableName() string {
	return "deposit"
}

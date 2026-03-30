package models

// Author: Benjamin Stonestreet
// Created: 2026-02-02

import (
	"time"

	"github.com/shopspring/decimal"
)

// Invoice represents a billing record for a specific customer. It tracks the amount
// charged, the current status of the payment, any associated fees, and the cryptocurrency used.
type Invoice struct {
	InvoiceID            string          `gorm:"primaryKey;type:varchar(36)" json:"invoice_id"`
	InvoiceDateTime      time.Time       `gorm:"not null;autoCreateTime" json:"invoice_date_time"`
	InvoiceAmountCharged decimal.Decimal `gorm:"not null;type:decimal(16,4)" json:"invoice_amount_charged"`
	InvoiceStatus        string          `gorm:"not null" json:"invoice_status"`
	InvoiceFeeAmount     decimal.Decimal `gorm:"not null;type:decimal(16,4)" json:"invoice_fee_amount"`
	InvoiceFeeStatus     string          `gorm:"not null" json:"invoice_fee_status"`
	InvoiceCryptoType    string          `gorm:"not null" json:"invoice_crypto_type"`
	InvoiceCustomerID    string          `gorm:"not null;type:varchar(36)" json:"invoice_customer_id"`
	Customer             Customer        `gorm:"foreignKey:InvoiceCustomerID;references:CustomerID" json:"customer,omitempty"`
}

// TableName overrides the default GORM table name to "invoice".
func (Invoice) TableName() string {
	return "invoice"
}

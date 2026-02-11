package models

import "github.com/shopspring/decimal"

type Invoice struct {
	InvoiceID            string          `gorm:"primaryKey;type:varchar(36)" json:"invoice_id"`
	InvoiceAmountCharged decimal.Decimal `gorm:"not null;type:decimal(16,4)" json:"invoice_amount_charged"`
	InvoiceStatus        string          `gorm:"not null" json:"invoice_status"`
	InvoiceFeeAmount     decimal.Decimal `gorm:"not null;type:decimal(16,4)" json:"invoice_fee_amount"`
	InvoiceFeeStatus     string          `gorm:"not null" json:"invoice_fee_status"`
	InvoiceCryptoType    string          `gorm:"not null" json:"invoice_crypto_type"`
	InvoiceCustomerID    string          `gorm:"not null;type:varchar(36)" json:"invoice_customer_id"`
	Customer             Customer        `gorm:"foreignKey:InvoiceCustomerID;references:CustomerID" json:"customer,omitempty"`
}

func (Invoice) TableName() string {
	return "invoice"
}

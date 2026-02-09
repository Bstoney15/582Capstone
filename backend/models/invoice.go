package models

type Invoice struct {
	InvoiceID            string   `gorm:"primaryKey" json:"invoice_id"`
	InvoiceAmountCharged float64  `gorm:"not null" json:"invoice_amount_charged"`
	InvoiceStatus        string   `gorm:"not null" json:"invoice_status"`
	InvoiceFeeAmount     string   `gorm:"not null" json:"invoice_fee_amount"`
	InvoiceFeeStatus     string   `gorm:"not null" json:"invoice_fee_status"`
	InvoiceCryptoType    string   `gorm:"not null" json:"invoice_crypto_type"`
	InvoiceCustomerID    string   `gorm:"not null" json:"invoice_customer_id"`
	Customer             Customer `gorm:"foreignKey:InvoiceCustomerID;references:CustomerID" json:"customer,omitempty"`
}

func (Invoice) TableName() string {
	return "invoice"
}

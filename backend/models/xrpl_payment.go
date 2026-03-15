package models

import "time"

type XRPLPayment struct {
	TxHash           string    `gorm:"primaryKey;type:varchar(128)" json:"tx_hash"`
	Destination      string    `gorm:"not null" json:"destination"`
	AmountDrops      string    `gorm:"not null" json:"amount_drops"`
	DestinationTag   *uint32   `json:"destination_tag,omitempty"`
	LedgerIndex      int64     `gorm:"not null" json:"ledger_index"`
	InvoiceID        *string   `gorm:"type:varchar(36)" json:"invoice_id,omitempty"`
	ProcessedAt      time.Time `gorm:"not null" json:"processed_at"`
	WalletMerchantID string    `gorm:"not null;type:varchar(36)" json:"wallet_merchant_id"`
}

func (XRPLPayment) TableName() string {
	return "xrpl_payment"
}
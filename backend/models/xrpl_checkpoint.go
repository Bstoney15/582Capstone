package models

import "time"

type XRPLCheckpoint struct {
	Account         string    `gorm:"primaryKey;type:varchar(128)" json:"account"`
	LastLedgerIndex int64     `gorm:"not null" json:"last_ledger_index"`
	UpdatedAt       time.Time `gorm:"not null" json:"updated_at"`
}

func (XRPLCheckpoint) TableName() string {
	return "xrpl_checkpoint"
}
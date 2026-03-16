package models

// Author: Benjamin Stonestreet
// Created: 2024-02-02

import "time"

// XRPLCheckpoint tracks the last processed ledger index for a given XRP account
// to ensure the reconciler resumes from the correct point after a restart.
type XRPLCheckpoint struct {
	Account         string    `gorm:"primaryKey;type:varchar(128)" json:"account"`
	LastLedgerIndex int64     `gorm:"not null" json:"last_ledger_index"`
	UpdatedAt       time.Time `gorm:"not null" json:"updated_at"`
}

// TableName overrides the default GORM table name to "xrpl_checkpoint".
func (XRPLCheckpoint) TableName() string {
	return "xrpl_checkpoint"
}
package models

type Deposit struct {
	DepositID         string   `gorm:"primaryKey" json:"deposit_id"`
	DepositAmount     float64  `gorm:"not null" json:"deposit_amount"`
	DepositStatus     string   `gorm:"not null" json:"deposit_status"`
	DepositFeeAmount  float64  `gorm:"not null" json:"deposit_fee_amount"`
	DepositFeeStatus  string   `gorm:"not null" json:"deposit_fee_status"`
	DepositCryptoType string   `gorm:"not null" json:"deposit_crypto_type"`
	DepositCustomerID string   `gorm:"not null" json:"deposit_customer_id"`
	Customer          Customer `gorm:"foreignKey:DepositCustomerID;references:CustomerID" json:"customer,omitempty"`
}

func (Deposit) TableName() string {
	return "deposit"
}

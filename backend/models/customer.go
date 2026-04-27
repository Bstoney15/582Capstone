// customer.go – GORM model representing a customer associated with a merchant.
package models

// Author: Benjamin Stonestreet
// Created: 2026-02-09

// Customer represents a customer of a merchant, storing their basic contact info
// and tying them to a specific merchant via CustomerMerchantID.
type Customer struct {
	CustomerID         string   `gorm:"primaryKey;type:varchar(36)" json:"customer_id"`
	CustomerMerchantID string   `gorm:"not null;type:varchar(36)" json:"customer_merchant_id"`
	Merchant           Merchant `gorm:"foreignKey:CustomerMerchantID;references:MerchantID" json:"merchant,omitempty"`
	CustomerFirstName  string   `json:"customer_first_name"`
	CustomerLastName   string   `json:"customer_last_name"`
	CustomerEmail      string   `json:"customer_email"`
}

// TableName overrides the default GORM table name to specifically be "merchant_customers".
func (Customer) TableName() string {
	return "merchant_customers"
}

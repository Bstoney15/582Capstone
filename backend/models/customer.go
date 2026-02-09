package models

type Customer struct {
	CustomerID         string   `gorm:"primaryKey" json:"customer_id"`
	CustomerMerchantID string   `gorm:"not null" json:"customer_merchant_id"`
	Merchant           Merchant `gorm:"foreignKey:CustomerMerchantID;references:MerchantID" json:"merchant,omitempty"`
	CustomerFirstName  string   `json:"customer_first_name"`
	CustomerLastName   string   `json:"customer_last_name"`
	CustomerEmail      string   `json:"customer_email"`
}

func (Customer) TableName() string {
	return "merchant_customers"
}

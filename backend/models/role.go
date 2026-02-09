package models

const (
	RoleAdmin     = "Admin"
	RoleDeveloper = "Developer"
)

type Role struct {
	RoleID         string   `gorm:"primaryKey" json:"role_id"`
	RoleMerchantID string   `gorm:"not null" json:"role_merchant_id"`
	Merchant       Merchant `gorm:"foreignKey:RoleMerchantID;references:MerchantID" json:"merchant,omitempty"`
	RoleName       string   `gorm:"not null" json:"role_name"`
}

func (Role) TableName() string {
	return "role"
}

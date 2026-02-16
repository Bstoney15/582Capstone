package models

const (
	RoleAdmin     = "Admin"
	RoleDeveloper = "Developer"
)

type Role struct {
	RoleID         string   `gorm:"primaryKey;type:varchar(36)" json:"role_id"`

	RoleMerchantID string   `gorm:"not null;type:varchar(36)" json:"role_merchant_id"`
	Merchant       Merchant `gorm:"foreignKey:RoleMerchantID;references:MerchantID" json:"merchant,omitempty"`

	RoleUserID     string   `gorm:"not null;type:varchar(36)" json:"role_user_id"`
	User           User     `gorm:"foreignKey:RoleUserID;references:UserID" json:"user,omitempty"`
	
	RoleName       string   `gorm:"not null" json:"role_name"`
}

func (Role) TableName() string {
	return "role"
}

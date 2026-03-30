package models

// Author: Benjamin Stonestreet
// Created: 2026-02-02

// Role definitions for authorization boundaries within a merchant organization.
const (
	RoleAdmin     = "Admin"
	RoleDeveloper = "Developer"
	RoleOwner     = "Owner"
)

// Role maps a User to a Merchant with a specific authorization level.
type Role struct {
	RoleID string `gorm:"primaryKey;type:varchar(36)" json:"role_id"`

	RoleMerchantID string   `gorm:"not null;type:varchar(36)" json:"role_merchant_id"`
	Merchant       Merchant `gorm:"foreignKey:RoleMerchantID;references:MerchantID" json:"merchant,omitempty"`

	RoleUserID string `gorm:"not null;type:varchar(36)" json:"role_user_id"`
	User       User   `gorm:"foreignKey:RoleUserID;references:UserID" json:"user,omitempty"`

	RoleName string `gorm:"not null" json:"role_name"`
}

// TableName sets the table name for Role.
func (Role) TableName() string {
	return "role"
}

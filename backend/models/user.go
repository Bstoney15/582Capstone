package models

// Author: Benjamin Stonestreet
// Created: 2024-02-02

// User represents a registered user who can hold roles across multiple merchants.
type User struct {
	UserID           string `gorm:"primaryKey;type:varchar(36)" json:"user_id"`
	UserUsername     string `gorm:"unique;not null" json:"user_username"`
	UserFirstName    string `gorm:"not null" json:"user_first_name"`
	UserLastName     string `gorm:"not null" json:"user_last_name"`
	UserPasswordHash string `gorm:"not null" json:"user_password_hash"`
	UserStatus       string `gorm:"not null" json:"user_status"`
}

// TableName sets the table name for User.
func (User) TableName() string {
	return "users"
}

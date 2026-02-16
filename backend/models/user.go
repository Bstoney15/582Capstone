package models

type User struct {
	UserID           string `gorm:"primaryKey;type:varchar(36)" json:"user_id"`
	UserUsername     string `gorm:"not null" json:"user_username"`
	UserFirstName    string `gorm:"not null" json:"user_first_name"`
	UserLastName     string `gorm:"not null" json:"user_last_name"`
	UserPasswordHash string `gorm:"not null" json:"user_password_hash"`
	UserStatus       string `gorm:"not null" json:"user_status"`
}

func (User) TableName() string {
	return "users"
}

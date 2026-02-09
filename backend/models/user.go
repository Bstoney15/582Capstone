package models

type User struct {
	UserID           string `gorm:"primaryKey" json:"user_id"`
	UserUsername     string `gorm:"not null" json:"user_username"`
	UserFirstName    string `gorm:"not null" json:"user_first_name"`
	UserLastName     string `gorm:"not null" json:"user_last_name"`
	UserPasswordHash string `gorm:"not null" json:"user_password_hash"`
	UserStatus       string `gorm:"not null" json:"user_status"`
	UserRoleID       string `gorm:"not null" json:"user_role_id"`
	Role             Role   `gorm:"foreignKey:UserRoleID;references:RoleID" json:"role,omitempty"`
}

func (User) TableName() string {
	return "user"
}

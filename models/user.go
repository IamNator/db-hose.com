package models

import "time"

type User struct {
	ID           string    `json:"id" gorm:"primary_key"`
	PasswordSalt string    `json:"password_salt" gorm:"not null"`
	Password     string    `json:"password" gorm:"not null"`
	Email        string    `json:"email" gorm:"unique"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChangePasswordData struct {
	Email           string `json:"email"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

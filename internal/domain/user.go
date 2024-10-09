package domain

import "time"

type User struct {
	ID           string    `json:"id" gorm:"primary_key"`
	PasswordSalt string    `json:"password_salt" gorm:"not null"`
	Password     string    `json:"password" gorm:"not null"`
	Email        string    `json:"email" gorm:"unique"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) Update(user User) {
	if user.PasswordSalt != "" {
		u.PasswordSalt = user.PasswordSalt
	}

	if user.Password != "" {
		u.Password = user.Password
	}

	if user.Email != "" {
		u.Email = user.Email
	}

	u.UpdatedAt = time.Now()
}

package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	PasswordHash   string    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastLogin      time.Time `json:"last_login"`
	DeviceToken    string    `json:"-"`
	LastLoginToken time.Time `json:"-"` // Used to validate token timestamps

}

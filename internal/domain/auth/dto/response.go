package dto

import "time"

// AuthResponse represents the response body for a successful login.
type AuthResponse struct {
	Token string `json:"token"`
}

// ProfileResponse represents the response body for user profile information.
type ProfileResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	LastLogin time.Time `json:"last_login"`
}

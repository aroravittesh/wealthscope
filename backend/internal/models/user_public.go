package models

import "time"

// UserPublic is a safe representation of a user (no password hash).
type UserPublic struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	RiskPreference string    `json:"risk_preference"`
	CreatedAt      time.Time `json:"created_at"`
}

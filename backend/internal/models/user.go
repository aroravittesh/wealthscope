package models

import "time"

type User struct {
	ID              string
	Email           string
	PasswordHash    string
	Role            string
	RiskPreference  string
	CreatedAt       time.Time
}

package repository

import "time"

type RefreshToken struct {
	ID         string
	UserID     string
	Token      string
	LastUsedAt time.Time
	ExpiresAt  time.Time
}

type RefreshTokenRepository interface {
	Create(token *RefreshToken) error
	Find(token string) (*RefreshToken, error)
	UpdateLastUsed(token string, t time.Time) error
	Delete(token string) error
	DeleteByUser(userID string) error
}

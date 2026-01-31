package repository

import (
	"database/sql"
	"errors"
	"time"
)

type RefreshTokenRepositoryPG struct {
	DB *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepositoryPG {
	return &RefreshTokenRepositoryPG{DB: db}
}

func (r *RefreshTokenRepositoryPG) Create(t *RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, last_used_at, expires_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.DB.Exec(query, t.UserID, t.Token, t.LastUsedAt, t.ExpiresAt)
	return err
}

func (r *RefreshTokenRepositoryPG) Find(token string) (*RefreshToken, error) {
	query := `
		SELECT id, user_id, token, last_used_at, expires_at
		FROM refresh_tokens
		WHERE token = $1
	`
	row := r.DB.QueryRow(query, token)

	var rt RefreshToken
	err := row.Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.LastUsedAt, &rt.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("refresh token not found")
	}
	return &rt, err
}

func (r *RefreshTokenRepositoryPG) UpdateLastUsed(token string, t time.Time) error {
	_, err := r.DB.Exec(
		`UPDATE refresh_tokens SET last_used_at = $1 WHERE token = $2`,
		t, token,
	)
	return err
}

func (r *RefreshTokenRepositoryPG) Delete(token string) error {
	_, err := r.DB.Exec(`DELETE FROM refresh_tokens WHERE token = $1`, token)
	return err
}

func (r *RefreshTokenRepositoryPG) DeleteByUser(userID string) error {
	_, err := r.DB.Exec(`DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}

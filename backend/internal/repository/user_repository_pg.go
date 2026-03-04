package repository

import (
	"database/sql"
	"errors"

	"wealthscope-backend/internal/models"
)

type UserRepositoryPG struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepositoryPG {
	return &UserRepositoryPG{DB: db}
}

func (r *UserRepositoryPG) Create(user *models.User) error {
	query := `
		INSERT INTO users (email, password_hash, role, risk_preference)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.DB.Exec(
		query,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.RiskPreference,
	)
	return err
}

func (r *UserRepositoryPG) FindByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, risk_preference, created_at
		FROM users
		WHERE email = $1
	`

	row := r.DB.QueryRow(query, email)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.RiskPreference,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}

	return &user, err
}

func (r *UserRepositoryPG) FindByID(id string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, risk_preference, created_at
		FROM users
		WHERE id = $1
	`

	row := r.DB.QueryRow(query, id)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.RiskPreference,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}

	return &user, err
}

func (r *UserRepositoryPG) UpdatePassword(userID string, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1
		WHERE id = $2
	`
	_, err := r.DB.Exec(query, passwordHash, userID)
	return err
}

func (r *UserRepositoryPG) UpdateRiskPreference(userID string, riskPreference string) error {
	query := `
		UPDATE users
		SET risk_preference = $1
		WHERE id = $2
	`
	_, err := r.DB.Exec(query, riskPreference, userID)
	return err
}

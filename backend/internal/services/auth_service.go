package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/repository"
)

type AuthService struct {
	UserRepo         repository.UserRepository
	RefreshTokenRepo repository.RefreshTokenRepository
}

/* ======================
   PASSWORD HELPERS
====================== */

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func checkPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

/* ======================
   JWT ACCESS TOKEN
====================== */

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func generateAccessToken(user *models.User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

/* ======================
   REFRESH TOKEN
====================== */

func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

/* ======================
   AUTH BUSINESS LOGIC
====================== */

func (s *AuthService) Register(
	email string,
	password string,
	riskPreference string,
) error {

	_, err := s.UserRepo.FindByEmail(email)
	if err == nil {
		return errors.New("email already registered")
	}

	hash, err := hashPassword(password)
	if err != nil {
		return err
	}

	user := &models.User{
		Email:          email,
		PasswordHash:   hash,
		Role:           "USER",
		RiskPreference: riskPreference,
	}

	return s.UserRepo.Create(user)
}

func (s *AuthService) Login(
	email string,
	password string,
) (string, string, error) {

	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if err := checkPassword(password, user.PasswordHash); err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if !user.IsEmailVerified {
		return "", "", errors.New("email not verified")
	}

	// Generate refresh token
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return "", "", err
	}

	now := time.Now().UTC()

	rt := &repository.RefreshToken{
		UserID:     user.ID,
		Token:      refreshToken,
		LastUsedAt: now,
		ExpiresAt:  now.Add(1 * time.Hour),
	}

	if err := s.RefreshTokenRepo.Create(rt); err != nil {
		return "", "", err
	}

	// Generate access token
	accessToken, err := generateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) RefreshAccessToken(userID string) (string, error) {
	user, err := s.UserRepo.FindByID(userID)
	if err != nil {
		return "", err
	}
	return generateAccessToken(user)
}

func GenerateNewRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *AuthService) ChangePassword(
	userID string,
	oldPassword string,
	newPassword string,
) error {

	user, err := s.UserRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	if err := checkPassword(oldPassword, user.PasswordHash); err != nil {
		return errors.New("invalid old password")
	}

	// Hash new password
	newHash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	if err := s.UserRepo.UpdatePassword(userID, newHash); err != nil {
		return err
	}

	// Invalidate all refresh tokens (force re-login everywhere)
	_ = s.RefreshTokenRepo.DeleteByUser(userID)

	return nil
}

package services

import (
	"errors"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"

	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/repository"
)

type AuthService struct {
	UserRepo repository.UserRepository
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func checkPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

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
) (string, error) {

	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := checkPassword(password, user.PasswordHash); err != nil {
		return "", errors.New("invalid credentials")
	}

	if !user.IsEmailVerified {
		return "", errors.New("email not verified")
	}

	return generateAccessToken(user)
}

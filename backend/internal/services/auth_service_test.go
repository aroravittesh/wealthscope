package services

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/repository"
)

type fakeUserRepository struct {
	findByEmailFn       func(email string) (*models.User, error)
	createFn            func(user *models.User) error
	findByIDFn          func(id string) (*models.User, error)
	updatePasswordFn    func(userID string, passwordHash string) error
	updateRiskPrefFn    func(userID string, riskPreference string) error
	createCalls         int
	updatePasswordCalls int
}

func (f *fakeUserRepository) Create(user *models.User) error {
	f.createCalls++
	if f.createFn == nil {
		return nil
	}
	return f.createFn(user)
}

func (f *fakeUserRepository) FindByEmail(email string) (*models.User, error) {
	if f.findByEmailFn == nil {
		return nil, errors.New("not implemented")
	}
	return f.findByEmailFn(email)
}

func (f *fakeUserRepository) FindByID(id string) (*models.User, error) {
	if f.findByIDFn == nil {
		return nil, errors.New("not implemented")
	}
	return f.findByIDFn(id)
}

func (f *fakeUserRepository) UpdatePassword(userID string, passwordHash string) error {
	f.updatePasswordCalls++
	if f.updatePasswordFn == nil {
		return nil
	}
	return f.updatePasswordFn(userID, passwordHash)
}

func (f *fakeUserRepository) UpdateRiskPreference(userID string, riskPreference string) error {
	if f.updateRiskPrefFn == nil {
		return nil
	}
	return f.updateRiskPrefFn(userID, riskPreference)
}

func (f *fakeUserRepository) ListAllPublic() ([]models.UserPublic, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeUserRepository) UpdateRole(userID string, role string) error {
	return errors.New("not implemented")
}

type fakeRefreshTokenRepository struct {
	findFn         func(token string) (*repository.RefreshToken, error)
	createFn       func(t *repository.RefreshToken) error
	deleteFn       func(token string) error
	deleteByUserFn func(userID string) error

	createCalls int
	deleteCalls int

	lastCreatedToken  *repository.RefreshToken
	lastDeletedToken  string
	lastDeletedByUser string
}

func (f *fakeRefreshTokenRepository) Create(t *repository.RefreshToken) error {
	f.createCalls++
	f.lastCreatedToken = t
	if f.createFn == nil {
		return nil
	}
	return f.createFn(t)
}

func (f *fakeRefreshTokenRepository) Find(token string) (*repository.RefreshToken, error) {
	if f.findFn == nil {
		return nil, errors.New("not implemented")
	}
	return f.findFn(token)
}

func (f *fakeRefreshTokenRepository) UpdateLastUsed(token string, t time.Time) error {
	return nil
}

func (f *fakeRefreshTokenRepository) Delete(token string) error {
	f.deleteCalls++
	f.lastDeletedToken = token
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(token)
}

func (f *fakeRefreshTokenRepository) DeleteByUser(userID string) error {
	f.lastDeletedByUser = userID
	if f.deleteByUserFn == nil {
		return nil
	}
	return f.deleteByUserFn(userID)
}

func TestAuthService_Register_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	password := "P@ssw0rd-123"
	var created *models.User

	userRepo := &fakeUserRepository{
		findByEmailFn: func(email string) (*models.User, error) {
			return nil, errors.New("user not found")
		},
		createFn: func(user *models.User) error {
			created = user
			return nil
		},
	}
	rtRepo := &fakeRefreshTokenRepository{}

	svc := &AuthService{UserRepo: userRepo, RefreshTokenRepo: rtRepo}
	if err := svc.Register("test@example.com", password, "MEDIUM"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if created == nil {
		t.Fatalf("expected user to be created")
	}
	if created.Role != "USER" {
		t.Fatalf("expected role USER, got %q", created.Role)
	}
	if created.Email != "test@example.com" {
		t.Fatalf("unexpected email: %q", created.Email)
	}
	if created.RiskPreference != "MEDIUM" {
		t.Fatalf("unexpected risk preference: %q", created.RiskPreference)
	}
	if created.PasswordHash == "" || created.PasswordHash == password {
		t.Fatalf("expected a hashed password, got %q", created.PasswordHash)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(created.PasswordHash), []byte(password)); err != nil {
		t.Fatalf("expected hash to match password, got error: %v", err)
	}
	if userRepo.createCalls != 1 {
		t.Fatalf("expected 1 create call, got %d", userRepo.createCalls)
	}
}

func TestAuthService_Register_EmailAlreadyRegistered(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	userRepo := &fakeUserRepository{
		findByEmailFn: func(email string) (*models.User, error) {
			return &models.User{
				ID:           "u1",
				Email:        email,
				PasswordHash: "hash",
				Role:         "USER",
			}, nil
		},
	}
	rtRepo := &fakeRefreshTokenRepository{}

	svc := &AuthService{UserRepo: userRepo, RefreshTokenRepo: rtRepo}
	err := svc.Register("test@example.com", "password", "LOW")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "email already registered" {
		t.Fatalf("expected email already registered error, got %v", err)
	}
	if userRepo.createCalls != 0 {
		t.Fatalf("expected create not to be called, got %d", userRepo.createCalls)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	password := "P@ssw0rd-123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		t.Fatalf("bcrypt setup failed: %v", err)
	}

	user := &models.User{
		ID:           "u1",
		Email:        "test@example.com",
		PasswordHash: string(hash),
		Role:         "USER",
	}

	userRepo := &fakeUserRepository{
		findByEmailFn: func(email string) (*models.User, error) {
			if email != user.Email {
				t.Fatalf("unexpected email arg: %q", email)
			}
			return user, nil
		},
	}

	rtRepo := &fakeRefreshTokenRepository{
		createFn: func(t *repository.RefreshToken) error {
			return nil
		},
	}

	svc := &AuthService{UserRepo: userRepo, RefreshTokenRepo: rtRepo}

	accessToken, refreshToken, err := svc.Login(user.Email, password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if accessToken == "" {
		t.Fatalf("expected non-empty access token")
	}
	if refreshToken == "" {
		t.Fatalf("expected non-empty refresh token")
	}
	if rtRepo.lastCreatedToken == nil {
		t.Fatalf("expected refresh token repo to receive a Create() call")
	}
	if refreshToken != rtRepo.lastCreatedToken.Token {
		t.Fatalf("expected returned refresh token to match created token")
	}
	if rtRepo.lastCreatedToken.UserID != user.ID {
		t.Fatalf("expected refresh token user_id %q, got %q", user.ID, rtRepo.lastCreatedToken.UserID)
	}

	// Validate JWT claims.
	parsed, err := jwt.ParseWithClaims(accessToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("expected valid jwt access token, got err=%v valid=%v", err, parsed.Valid)
	}
	claims := parsed.Claims.(*Claims)
	if claims.UserID != user.ID {
		t.Fatalf("expected claims.UserID %q, got %q", user.ID, claims.UserID)
	}
	if claims.Role != user.Role {
		t.Fatalf("expected claims.Role %q, got %q", user.Role, claims.Role)
	}

	now := time.Now().UTC()
	if rtRepo.lastCreatedToken.ExpiresAt.Before(now.Add(-1 * time.Second)) {
		t.Fatalf("expected refresh token ExpiresAt to be in the future")
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	userRepo := &fakeUserRepository{
		findByEmailFn: func(email string) (*models.User, error) {
			return nil, errors.New("user not found")
		},
	}
	rtRepo := &fakeRefreshTokenRepository{}
	svc := &AuthService{UserRepo: userRepo, RefreshTokenRepo: rtRepo}

	_, _, err := svc.Login("test@example.com", "wrong-password")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "invalid credentials" {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
	if rtRepo.createCalls != 0 {
		t.Fatalf("expected refresh token Create() not to be called, got %d", rtRepo.createCalls)
	}
}

func TestAuthService_RefreshAccessToken_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	oldToken := "old-refresh-token"
	future := time.Now().UTC().Add(10 * time.Minute)
	user := &models.User{
		ID:           "u1",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         "USER",
	}

	userRepo := &fakeUserRepository{
		findByIDFn: func(id string) (*models.User, error) {
			if id != user.ID {
				t.Fatalf("unexpected userID: %q", id)
			}
			return user, nil
		},
	}

	rtRepo := &fakeRefreshTokenRepository{
		findFn: func(token string) (*repository.RefreshToken, error) {
			if token != oldToken {
				t.Fatalf("unexpected old token: %q", token)
			}
			return &repository.RefreshToken{
				UserID:    user.ID,
				Token:     oldToken,
				ExpiresAt: future,
			}, nil
		},
	}

	svc := &AuthService{UserRepo: userRepo, RefreshTokenRepo: rtRepo}
	accessToken, refreshToken, err := svc.RefreshAccessToken(oldToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if accessToken == "" || refreshToken == "" {
		t.Fatalf("expected tokens to be non-empty")
	}

	if rtRepo.lastDeletedToken != oldToken {
		t.Fatalf("expected Delete(oldToken) to be called, got %q", rtRepo.lastDeletedToken)
	}
	if rtRepo.lastCreatedToken == nil {
		t.Fatalf("expected Create() to be called for rotated refresh token")
	}
	if refreshToken != rtRepo.lastCreatedToken.Token {
		t.Fatalf("expected returned refresh token to match created token")
	}

	parsed, err := jwt.ParseWithClaims(accessToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("expected valid jwt access token, got err=%v valid=%v", err, parsed.Valid)
	}
	claims := parsed.Claims.(*Claims)
	if claims.UserID != user.ID {
		t.Fatalf("expected claims.UserID %q, got %q", user.ID, claims.UserID)
	}
}

func TestAuthService_RefreshAccessToken_Expired(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	oldToken := "old-refresh-token"
	userRepo := &fakeUserRepository{}
	rtRepo := &fakeRefreshTokenRepository{
		findFn: func(token string) (*repository.RefreshToken, error) {
			return &repository.RefreshToken{
				UserID:    "u1",
				Token:     oldToken,
				ExpiresAt: time.Now().UTC().Add(-1 * time.Minute),
			}, nil
		},
	}

	svc := &AuthService{UserRepo: userRepo, RefreshTokenRepo: rtRepo}
	_, _, err := svc.RefreshAccessToken(oldToken)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "refresh token expired" {
		t.Fatalf("expected refresh token expired error, got %v", err)
	}
	if rtRepo.createCalls != 0 || rtRepo.deleteCalls != 0 {
		t.Fatalf("expected no rotation calls on expired token (createCalls=%d deleteCalls=%d)", rtRepo.createCalls, rtRepo.deleteCalls)
	}
}

func TestAuthService_ChangePassword_Success(t *testing.T) {
	passwordOld := "old-password"
	passwordNew := "new-password"

	oldHash, err := bcrypt.GenerateFromPassword([]byte(passwordOld), 12)
	if err != nil {
		t.Fatalf("bcrypt setup failed: %v", err)
	}

	userRepo := &fakeUserRepository{
		findByIDFn: func(id string) (*models.User, error) {
			return &models.User{
				ID:           id,
				PasswordHash: string(oldHash),
			}, nil
		},
		updatePasswordFn: func(userID string, passwordHash string) error {
			// Ensure UpdatePassword gets a new hash for new password.
			if userID != "u1" {
				t.Fatalf("unexpected userID: %q", userID)
			}
			if passwordHash == "" || passwordHash == passwordNew {
				t.Fatalf("expected a hashed password")
			}
			if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(passwordNew)); err != nil {
				t.Fatalf("expected updated hash to match new password: %v", err)
			}
			return nil
		},
	}

	rtRepo := &fakeRefreshTokenRepository{}
	svc := &AuthService{UserRepo: userRepo, RefreshTokenRepo: rtRepo}

	if err := svc.ChangePassword("u1", passwordOld, passwordNew); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userRepo.updatePasswordCalls != 1 {
		t.Fatalf("expected UpdatePassword to be called once, got %d", userRepo.updatePasswordCalls)
	}
	if rtRepo.lastDeletedByUser != "u1" {
		t.Fatalf("expected DeleteByUser(u1) to be called, got %q", rtRepo.lastDeletedByUser)
	}
}

func TestAuthService_ChangePassword_InvalidOldPassword(t *testing.T) {
	passwordOld := "old-password"
	passwordNew := "new-password"
	passwordWrong := "wrong-old-password"

	oldHash, err := bcrypt.GenerateFromPassword([]byte(passwordOld), 12)
	if err != nil {
		t.Fatalf("bcrypt setup failed: %v", err)
	}

	userRepo := &fakeUserRepository{
		findByIDFn: func(id string) (*models.User, error) {
			return &models.User{
				ID:           id,
				PasswordHash: string(oldHash),
			}, nil
		},
	}

	rtRepo := &fakeRefreshTokenRepository{}
	svc := &AuthService{UserRepo: userRepo, RefreshTokenRepo: rtRepo}

	err = svc.ChangePassword("u1", passwordWrong, passwordNew)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "invalid old password" {
		t.Fatalf("expected invalid old password error, got %v", err)
	}
	if userRepo.updatePasswordCalls != 0 {
		t.Fatalf("expected UpdatePassword not to be called, got %d", userRepo.updatePasswordCalls)
	}
	if rtRepo.lastDeletedByUser != "" {
		t.Fatalf("expected DeleteByUser not to be called")
	}
}

func TestAuthService_ChangePassword_UserNotFound(t *testing.T) {
	userRepo := &fakeUserRepository{
		findByIDFn: func(id string) (*models.User, error) {
			return nil, errors.New("user not found")
		},
	}
	rtRepo := &fakeRefreshTokenRepository{}
	svc := &AuthService{UserRepo: userRepo, RefreshTokenRepo: rtRepo}

	err := svc.ChangePassword("missing-user", "old-password", "new-password")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "user not found" {
		t.Fatalf("expected user not found error, got %v", err)
	}
	if userRepo.updatePasswordCalls != 0 {
		t.Fatalf("expected UpdatePassword not to be called")
	}
}

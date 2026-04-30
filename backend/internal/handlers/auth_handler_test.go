package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/repository"
	"wealthscope-backend/internal/services"
)

type fakeUserRepo struct {
	findByEmailFn func(email string) (*models.User, error)
	createFn      func(user *models.User) error
}

func (f *fakeUserRepo) Create(user *models.User) error {
	if f.createFn == nil {
		return nil
	}
	return f.createFn(user)
}

func (f *fakeUserRepo) FindByEmail(email string) (*models.User, error) {
	if f.findByEmailFn == nil {
		return nil, errors.New("not implemented")
	}
	return f.findByEmailFn(email)
}

func (f *fakeUserRepo) FindByID(id string) (*models.User, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeUserRepo) UpdatePassword(userID string, passwordHash string) error {
	return nil
}

func (f *fakeUserRepo) UpdateRiskPreference(userID string, riskPreference string) error {
	return nil
}

func (f *fakeUserRepo) ListAllPublic() ([]models.UserPublic, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeUserRepo) UpdateRole(userID string, role string) error {
	return nil
}

type fakeRefreshTokenRepo struct {
	createFn    func(t *repository.RefreshToken) error
	lastCreated *repository.RefreshToken
}

func (f *fakeRefreshTokenRepo) Create(t *repository.RefreshToken) error {
	f.lastCreated = t
	if f.createFn == nil {
		return nil
	}
	return f.createFn(t)
}

func (f *fakeRefreshTokenRepo) Find(token string) (*repository.RefreshToken, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeRefreshTokenRepo) UpdateLastUsed(token string, t time.Time) error {
	return nil
}

func (f *fakeRefreshTokenRepo) Delete(token string) error {
	return nil
}

func (f *fakeRefreshTokenRepo) DeleteByUser(userID string) error {
	return nil
}

func TestAuthHandler_Register_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	userRepo := &fakeUserRepo{
		findByEmailFn: func(email string) (*models.User, error) {
			return nil, errors.New("user not found")
		},
		createFn: func(user *models.User) error {
			return nil
		},
	}
	rtRepo := &fakeRefreshTokenRepo{}

	authSvc := &services.AuthService{
		UserRepo:         userRepo,
		RefreshTokenRepo: rtRepo,
	}
	h := NewAuthHandler(authSvc)

	body := []byte(`{"email":"test@example.com","password":"password123","risk_preference":"LOW"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Result().StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Result().StatusCode)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}
	if resp["message"] == "" {
		t.Fatalf("expected message field in response")
	}
}

func TestAuthHandler_Register_InvalidBody(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	userRepo := &fakeUserRepo{}
	rtRepo := &fakeRefreshTokenRepo{}
	authSvc := &services.AuthService{UserRepo: userRepo, RefreshTokenRepo: rtRepo}
	h := NewAuthHandler(authSvc)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte(`{"email":`)))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Result().StatusCode)
	}
	if got := rec.Body.String(); got == "" || !contains(got, "invalid request body") {
		t.Fatalf("expected invalid request body message, got %q", got)
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	password := "P@ssw0rd-123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		t.Fatalf("bcrypt setup failed: %v", err)
	}

	userRepo := &fakeUserRepo{
		findByEmailFn: func(email string) (*models.User, error) {
			if email != "test@example.com" {
				t.Fatalf("unexpected email arg: %q", email)
			}
			return &models.User{
				ID:           "u1",
				Email:        email,
				PasswordHash: string(hash),
				Role:         "USER",
			}, nil
		},
	}
	rtRepo := &fakeRefreshTokenRepo{}

	authSvc := &services.AuthService{
		UserRepo:         userRepo,
		RefreshTokenRepo: rtRepo,
	}
	h := NewAuthHandler(authSvc)

	body := []byte(`{"email":"test@example.com","password":"` + password + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Result().StatusCode)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}
	if resp["access_token"] == "" || resp["refresh_token"] == "" {
		t.Fatalf("expected access_token and refresh_token in response")
	}

	// Basic sanity check: JWT should be parseable with our secret.
	parsed, err := jwt.Parse(resp["access_token"], func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("expected parseable access token, got err=%v valid=%v", err, parsed.Valid)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	password := "P@ssw0rd-123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		t.Fatalf("bcrypt setup failed: %v", err)
	}

	userRepo := &fakeUserRepo{
		findByEmailFn: func(email string) (*models.User, error) {
			return &models.User{
				ID:           "u1",
				Email:        email,
				PasswordHash: string(hash),
				Role:         "USER",
			}, nil
		},
	}
	rtRepo := &fakeRefreshTokenRepo{}

	authSvc := &services.AuthService{
		UserRepo:         userRepo,
		RefreshTokenRepo: rtRepo,
	}
	h := NewAuthHandler(authSvc)

	body := []byte(`{"email":"test@example.com","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Result().StatusCode)
	}
}

func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && bytes.Contains([]byte(s), []byte(substr)))
}

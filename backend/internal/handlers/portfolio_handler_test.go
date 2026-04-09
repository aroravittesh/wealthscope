package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"wealthscope-backend/internal/middleware"
	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/repository"
	"wealthscope-backend/internal/services"
)

type fakePortfolioRepo struct {
	createFn     func(p *models.Portfolio) error
	getByIDFn    func(id string) (*models.Portfolio, error)
	getByUserFn  func(userID string) ([]models.Portfolio, error)
	updateNameFn func(id string, name string) error
	deleteFn     func(id string) error

	createCalls int
}

func (f *fakePortfolioRepo) Create(p *models.Portfolio) error {
	f.createCalls++
	if f.createFn == nil {
		return nil
	}
	return f.createFn(p)
}

func (f *fakePortfolioRepo) GetByUser(userID string) ([]models.Portfolio, error) {
	if f.getByUserFn == nil {
		return nil, nil
	}
	return f.getByUserFn(userID)
}

func (f *fakePortfolioRepo) GetByID(id string) (*models.Portfolio, error) {
	if f.getByIDFn == nil {
		return nil, errors.New("not implemented")
	}
	return f.getByIDFn(id)
}

func (f *fakePortfolioRepo) UpdateName(id string, name string) error {
	if f.updateNameFn == nil {
		return nil
	}
	return f.updateNameFn(id, name)
}

func (f *fakePortfolioRepo) Delete(id string) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(id)
}

var _ repository.PortfolioRepository = (*fakePortfolioRepo)(nil)

type summaryFakeHoldingRepo struct{}

func (summaryFakeHoldingRepo) CreateOrUpdate(h *models.Holding) error {
	return errors.New("not implemented")
}

func (summaryFakeHoldingRepo) GetByPortfolio(portfolioID string) ([]models.Holding, error) {
	return nil, nil
}

func (summaryFakeHoldingRepo) Delete(id string) error { return errors.New("not implemented") }

func (summaryFakeHoldingRepo) FindBySymbol(portfolioID, symbol string) (*models.Holding, error) {
	return nil, errors.New("not implemented")
}

func (summaryFakeHoldingRepo) GetByID(id string) (*models.Holding, error) {
	return nil, errors.New("not implemented")
}

func (summaryFakeHoldingRepo) UpdateByID(id string, quantity float64, avgPrice float64) error {
	return errors.New("not implemented")
}

var _ repository.HoldingRepository = (*summaryFakeHoldingRepo)(nil)

func TestPortfolioHandler_Create_Success(t *testing.T) {
	repo := &fakePortfolioRepo{
		createFn: func(p *models.Portfolio) error {
			p.ID = "p1"
			return nil
		},
	}
	svc := &services.PortfolioService{PortfolioRepo: repo, HoldingRepo: summaryFakeHoldingRepo{}}
	h := NewPortfolioHandler(svc)

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "u1")
	req := httptest.NewRequest(http.MethodPost, "/portfolios", bytes.NewReader([]byte(`{"name":"My Portfolio"}`))).WithContext(ctx)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Result().StatusCode)
	}

	var got models.Portfolio
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}
	if got.ID != "p1" || got.UserID != "u1" || got.Name != "My Portfolio" {
		t.Fatalf("unexpected portfolio response: %#v", got)
	}
}

func TestPortfolioHandler_Create_InvalidName(t *testing.T) {
	repo := &fakePortfolioRepo{}
	svc := &services.PortfolioService{PortfolioRepo: repo}
	h := NewPortfolioHandler(svc)

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "u1")
	req := httptest.NewRequest(http.MethodPost, "/portfolios", bytes.NewReader([]byte(`{"name":""}`))).WithContext(ctx)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Result().StatusCode)
	}
}

func TestPortfolioHandler_Rename_Unauthorized(t *testing.T) {
	repo := &fakePortfolioRepo{
		getByIDFn: func(id string) (*models.Portfolio, error) {
			return &models.Portfolio{ID: id, UserID: "other-user", Name: "X"}, nil
		},
	}
	svc := &services.PortfolioService{PortfolioRepo: repo}
	h := NewPortfolioHandler(svc)

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "u1")
	req := httptest.NewRequest(http.MethodPut, "/portfolios/p1/rename", bytes.NewReader([]byte(`{"name":"New"}`))).WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": "p1"})
	rec := httptest.NewRecorder()

	h.Rename(rec, req)

	if rec.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Result().StatusCode)
	}
}

func TestPortfolioHandler_Delete_Success(t *testing.T) {
	repo := &fakePortfolioRepo{
		getByIDFn: func(id string) (*models.Portfolio, error) {
			return &models.Portfolio{ID: id, UserID: "u1"}, nil
		},
		deleteFn: func(id string) error { return nil },
	}
	svc := &services.PortfolioService{PortfolioRepo: repo, HoldingRepo: summaryFakeHoldingRepo{}}
	h := NewPortfolioHandler(svc)

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "u1")
	req := httptest.NewRequest(http.MethodDelete, "/portfolios/p1", nil).WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": "p1"})
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Result().StatusCode)
	}
}

func TestPortfolioHandler_GetSummary_NotFound(t *testing.T) {
	repo := &fakePortfolioRepo{
		getByIDFn: func(id string) (*models.Portfolio, error) {
			return nil, sql.ErrNoRows
		},
	}
	svc := &services.PortfolioService{PortfolioRepo: repo, HoldingRepo: summaryFakeHoldingRepo{}}
	h := NewPortfolioHandler(svc)

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "u1")
	req := httptest.NewRequest(http.MethodGet, "/portfolios/p1/summary", nil).WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": "p1"})
	rec := httptest.NewRecorder()

	h.GetSummary(rec, req)

	if rec.Result().StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Result().StatusCode)
	}
}

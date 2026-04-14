package handlers

import (
	"bytes"
	"context"
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

type fakeHoldingRepo struct {
	addFn        func(h *models.Holding) error
	getByPortFn  func(portfolioID string) ([]models.Holding, error)
	deleteFn     func(id string) error
	getByIDFn    func(id string) (*models.Holding, error)
	updateByIDFn func(id string, quantity float64, avgPrice float64) error
}

func (f *fakeHoldingRepo) CreateOrUpdate(h *models.Holding) error {
	if f.addFn == nil {
		return nil
	}
	return f.addFn(h)
}

func (f *fakeHoldingRepo) GetByPortfolio(portfolioID string) ([]models.Holding, error) {
	if f.getByPortFn == nil {
		return nil, nil
	}
	return f.getByPortFn(portfolioID)
}

func (f *fakeHoldingRepo) Delete(id string) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(id)
}

func (f *fakeHoldingRepo) FindBySymbol(portfolioID, symbol string) (*models.Holding, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeHoldingRepo) GetByID(id string) (*models.Holding, error) {
	if f.getByIDFn == nil {
		return nil, errors.New("not implemented")
	}
	return f.getByIDFn(id)
}

func (f *fakeHoldingRepo) UpdateByID(id string, quantity float64, avgPrice float64) error {
	if f.updateByIDFn == nil {
		return nil
	}
	return f.updateByIDFn(id, quantity, avgPrice)
}

type fakePortfolioRepoForHoldingHandler struct {
	getByIDFn func(id string) (*models.Portfolio, error)
}

func (f *fakePortfolioRepoForHoldingHandler) Create(portfolio *models.Portfolio) error {
	return nil
}

func (f *fakePortfolioRepoForHoldingHandler) GetByUser(userID string) ([]models.Portfolio, error) {
	return nil, nil
}

func (f *fakePortfolioRepoForHoldingHandler) GetByID(id string) (*models.Portfolio, error) {
	if f.getByIDFn == nil {
		return nil, errors.New("not implemented")
	}
	return f.getByIDFn(id)
}

func (f *fakePortfolioRepoForHoldingHandler) UpdateName(id string, name string) error {
	return nil
}

func (f *fakePortfolioRepoForHoldingHandler) Delete(id string) error {
	return nil
}

var _ repository.HoldingRepository = (*fakeHoldingRepo)(nil)

func TestHoldingHandler_Add_Success(t *testing.T) {
	var gotHolding *models.Holding
	repo := &fakeHoldingRepo{
		addFn: func(h *models.Holding) error {
			gotHolding = h
			return nil
		},
	}
	portfolioRepo := &fakePortfolioRepoForHoldingHandler{
		getByIDFn: func(id string) (*models.Portfolio, error) {
			return &models.Portfolio{ID: id, UserID: "u1"}, nil
		},
	}
	svc := &services.HoldingService{Repo: repo, PortfolioRepo: portfolioRepo}
	h := &HoldingHandler{Service: svc}

	req := httptest.NewRequest(http.MethodPost, "/holdings", bytes.NewReader([]byte(`{
		"portfolio_id":"p1",
		"symbol":"aapl",
		"asset_type":"STOCK",
		"quantity":2,
		"avg_price":150.25
	}`)))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "u1"))
	rec := httptest.NewRecorder()

	h.Add(rec, req)

	if rec.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Result().StatusCode)
	}
	if gotHolding == nil {
		t.Fatalf("expected holding to be passed to service")
	}
	if gotHolding.Symbol != "AAPL" {
		t.Fatalf("expected symbol to be uppercased by service, got %q", gotHolding.Symbol)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}
	if resp["message"] == "" {
		t.Fatalf("expected message field")
	}
}

func TestHoldingHandler_Get_Success(t *testing.T) {
	repo := &fakeHoldingRepo{
		getByPortFn: func(portfolioID string) ([]models.Holding, error) {
			if portfolioID != "p1" {
				t.Fatalf("unexpected portfolioID: %q", portfolioID)
			}
			return []models.Holding{{ID: "h1", PortfolioID: "p1", Symbol: "AAPL", AssetType: "STOCK", Quantity: 2, AvgPrice: 150.25}}, nil
		},
	}
	portfolioRepo := &fakePortfolioRepoForHoldingHandler{
		getByIDFn: func(id string) (*models.Portfolio, error) {
			return &models.Portfolio{ID: id, UserID: "u1"}, nil
		},
	}
	svc := &services.HoldingService{Repo: repo, PortfolioRepo: portfolioRepo}
	h := &HoldingHandler{Service: svc}

	req := httptest.NewRequest(http.MethodGet, "/portfolios/p1/holdings", nil)
	req = mux.SetURLVars(req, map[string]string{"portfolio_id": "p1"})
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "u1"))
	rec := httptest.NewRecorder()

	h.Get(rec, req)

	if rec.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Result().StatusCode)
	}
	var got []models.Holding
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "h1" {
		t.Fatalf("unexpected holdings: %#v", got)
	}
}

func TestHoldingHandler_Delete_Error(t *testing.T) {
	repo := &fakeHoldingRepo{
		getByIDFn: func(id string) (*models.Holding, error) {
			return &models.Holding{ID: id, PortfolioID: "p1"}, nil
		},
		deleteFn: func(id string) error {
			return errors.New("delete failed")
		},
	}
	portfolioRepo := &fakePortfolioRepoForHoldingHandler{
		getByIDFn: func(id string) (*models.Portfolio, error) {
			return &models.Portfolio{ID: id, UserID: "u1"}, nil
		},
	}
	svc := &services.HoldingService{Repo: repo, PortfolioRepo: portfolioRepo}
	h := &HoldingHandler{Service: svc}

	req := httptest.NewRequest(http.MethodDelete, "/holdings/h1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "h1"})
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "u1"))
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Result().StatusCode)
	}
}

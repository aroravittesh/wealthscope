package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/repository"
	"wealthscope-backend/internal/services"
)

type fakeHoldingRepo struct {
	addFn       func(h *models.Holding) error
	getByPortFn func(portfolioID string) ([]models.Holding, error)
	deleteFn    func(id string) error
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

var _ repository.HoldingRepository = (*fakeHoldingRepo)(nil)

func TestHoldingHandler_Add_Success(t *testing.T) {
	var gotHolding *models.Holding
	repo := &fakeHoldingRepo{
		addFn: func(h *models.Holding) error {
			gotHolding = h
			return nil
		},
	}
	svc := &services.HoldingService{Repo: repo}
	h := &HoldingHandler{Service: svc}

	req := httptest.NewRequest(http.MethodPost, "/holdings", bytes.NewReader([]byte(`{
		"portfolio_id":"p1",
		"symbol":"aapl",
		"asset_type":"STOCK",
		"quantity":2,
		"avg_price":150.25
	}`)))
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
	svc := &services.HoldingService{Repo: repo}
	h := &HoldingHandler{Service: svc}

	req := httptest.NewRequest(http.MethodGet, "/portfolios/p1/holdings", nil)
	req = mux.SetURLVars(req, map[string]string{"portfolio_id": "p1"})
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
		deleteFn: func(id string) error {
			return errors.New("delete failed")
		},
	}
	svc := &services.HoldingService{Repo: repo}
	h := &HoldingHandler{Service: svc}

	req := httptest.NewRequest(http.MethodDelete, "/holdings/h1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "h1"})
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Result().StatusCode)
	}
}


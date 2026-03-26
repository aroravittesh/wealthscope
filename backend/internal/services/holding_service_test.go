package services

import (
	"errors"
	"testing"

	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/repository"
)

type fakeHoldingRepository struct {
	createOrUpdateFn func(h *models.Holding) error
	getByPortfolioFn func(portfolioID string) ([]models.Holding, error)
	deleteFn         func(id string) error
	findBySymbolFn   func(portfolioID, symbol string) (*models.Holding, error)

	createOrUpdateCalls int
	lastCreatedOrUpdated *models.Holding
}

func (f *fakeHoldingRepository) CreateOrUpdate(h *models.Holding) error {
	f.createOrUpdateCalls++
	f.lastCreatedOrUpdated = h
	if f.createOrUpdateFn == nil {
		return nil
	}
	return f.createOrUpdateFn(h)
}

func (f *fakeHoldingRepository) GetByPortfolio(portfolioID string) ([]models.Holding, error) {
	if f.getByPortfolioFn == nil {
		return nil, nil
	}
	return f.getByPortfolioFn(portfolioID)
}

func (f *fakeHoldingRepository) Delete(id string) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(id)
}

func (f *fakeHoldingRepository) FindBySymbol(portfolioID, symbol string) (*models.Holding, error) {
	if f.findBySymbolFn == nil {
		return nil, errors.New("not implemented")
	}
	return f.findBySymbolFn(portfolioID, symbol)
}

func TestHoldingService_AddHolding_UppercasesSymbol(t *testing.T) {
	repo := &fakeHoldingRepository{}
	svc := &HoldingService{Repo: repo}

	if err := svc.AddHolding("p1", "aapl", "STOCK", 2, 150.25); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repo.createOrUpdateCalls != 1 {
		t.Fatalf("expected 1 CreateOrUpdate call, got %d", repo.createOrUpdateCalls)
	}

	h := repo.lastCreatedOrUpdated
	if h == nil {
		t.Fatalf("expected holding to be passed to repo")
	}
	if h.PortfolioID != "p1" {
		t.Fatalf("expected portfolio_id %q, got %q", "p1", h.PortfolioID)
	}
	if h.Symbol != "AAPL" {
		t.Fatalf("expected symbol to be uppercased to %q, got %q", "AAPL", h.Symbol)
	}
	if h.AssetType != "STOCK" {
		t.Fatalf("expected asset_type %q, got %q", "STOCK", h.AssetType)
	}
	if h.Quantity != 2 {
		t.Fatalf("expected quantity %v, got %v", 2, h.Quantity)
	}
	if h.AvgPrice != 150.25 {
		t.Fatalf("expected avg_price %v, got %v", 150.25, h.AvgPrice)
	}
}

func TestHoldingService_GetHoldings_PassesThrough(t *testing.T) {
	expected := []models.Holding{
		{ID: "h1", PortfolioID: "p1", Symbol: "AAPL", AssetType: "STOCK", Quantity: 2, AvgPrice: 150.25},
	}
	repo := &fakeHoldingRepository{
		getByPortfolioFn: func(portfolioID string) ([]models.Holding, error) {
			if portfolioID != "p1" {
				t.Fatalf("unexpected portfolioID: %q", portfolioID)
			}
			return expected, nil
		},
	}

	svc := &HoldingService{Repo: repo}
	got, err := svc.GetHoldings("p1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 holding, got %d", len(got))
	}
	if got[0].ID != "h1" {
		t.Fatalf("expected holding id %q, got %q", "h1", got[0].ID)
	}
}

func TestHoldingService_DeleteHolding_PassesThrough(t *testing.T) {
	repo := &fakeHoldingRepository{
		deleteFn: func(id string) error {
			if id != "h1" {
				t.Fatalf("unexpected id: %q", id)
			}
			return nil
		},
	}

	svc := &HoldingService{Repo: repo}
	if err := svc.DeleteHolding("h1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

var _ repository.HoldingRepository = (*fakeHoldingRepository)(nil)


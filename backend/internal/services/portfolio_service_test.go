package services

import (
	"errors"
	"testing"

	"wealthscope-backend/internal/market"
	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/repository"
)

type stubUnitPrices struct {
	fn func(symbol string, avg float64) float64
}

func (s stubUnitPrices) UnitPrice(symbol string, avgPrice float64) float64 {
	if s.fn == nil {
		return avgPrice
	}
	return s.fn(symbol, avgPrice)
}

type fakePortfolioRepository struct {
	createFn     func(p *models.Portfolio) error
	getByUserFn  func(userID string) ([]models.Portfolio, error)
	getByIDFn    func(id string) (*models.Portfolio, error)
	updateNameFn func(id string, name string) error
	deleteFn     func(id string) error

	createCalls int
	lastCreated *models.Portfolio

	updateNameCalls int
	lastUpdatedID   string
	lastUpdatedName string

	deleteCalls   int
	lastDeletedID string
}

func (f *fakePortfolioRepository) Create(portfolio *models.Portfolio) error {
	f.createCalls++
	f.lastCreated = portfolio
	if f.createFn == nil {
		return nil
	}
	return f.createFn(portfolio)
}

func (f *fakePortfolioRepository) GetByUser(userID string) ([]models.Portfolio, error) {
	if f.getByUserFn == nil {
		return nil, nil
	}
	return f.getByUserFn(userID)
}

func (f *fakePortfolioRepository) GetByID(id string) (*models.Portfolio, error) {
	if f.getByIDFn == nil {
		return nil, errors.New("not implemented")
	}
	return f.getByIDFn(id)
}

func (f *fakePortfolioRepository) UpdateName(id string, name string) error {
	f.updateNameCalls++
	f.lastUpdatedID = id
	f.lastUpdatedName = name
	if f.updateNameFn == nil {
		return nil
	}
	return f.updateNameFn(id, name)
}

func (f *fakePortfolioRepository) Delete(id string) error {
	f.deleteCalls++
	f.lastDeletedID = id
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(id)
}

var _ repository.PortfolioRepository = (*fakePortfolioRepository)(nil)

func TestPortfolioService_Create_ValidatesName(t *testing.T) {
	repo := &fakePortfolioRepository{}
	svc := &PortfolioService{PortfolioRepo: repo}

	_, err := svc.Create("u1", "")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "portfolio name required" {
		t.Fatalf("expected validation error, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected repo.Create not to be called")
	}
}

func TestPortfolioService_Create_PassesThrough(t *testing.T) {
	repo := &fakePortfolioRepository{
		createFn: func(p *models.Portfolio) error {
			p.ID = "p1"
			return nil
		},
	}
	svc := &PortfolioService{PortfolioRepo: repo}

	p, err := svc.Create("u1", "My Portfolio")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p == nil || p.ID != "p1" {
		t.Fatalf("expected created portfolio to have ID p1, got %#v", p)
	}
	if repo.lastCreated == nil || repo.lastCreated.UserID != "u1" || repo.lastCreated.Name != "My Portfolio" {
		t.Fatalf("expected portfolio to be passed with user_id and name")
	}
}

func TestPortfolioService_Rename_Unauthorized(t *testing.T) {
	repo := &fakePortfolioRepository{
		getByIDFn: func(id string) (*models.Portfolio, error) {
			return &models.Portfolio{ID: id, UserID: "other-user", Name: "X"}, nil
		},
	}
	svc := &PortfolioService{PortfolioRepo: repo}

	err := svc.Rename("u1", "p1", "New Name")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
	if repo.updateNameCalls != 0 {
		t.Fatalf("expected UpdateName not to be called")
	}
}

func TestPortfolioService_Rename_Success(t *testing.T) {
	repo := &fakePortfolioRepository{
		getByIDFn: func(id string) (*models.Portfolio, error) {
			return &models.Portfolio{ID: id, UserID: "u1", Name: "X"}, nil
		},
	}
	svc := &PortfolioService{PortfolioRepo: repo}

	err := svc.Rename("u1", "p1", "New Name")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updateNameCalls != 1 {
		t.Fatalf("expected 1 UpdateName call, got %d", repo.updateNameCalls)
	}
	if repo.lastUpdatedID != "p1" || repo.lastUpdatedName != "New Name" {
		t.Fatalf("expected UpdateName(p1, New Name), got UpdateName(%q, %q)", repo.lastUpdatedID, repo.lastUpdatedName)
	}
}

func TestPortfolioService_Delete_Unauthorized(t *testing.T) {
	repo := &fakePortfolioRepository{
		getByIDFn: func(id string) (*models.Portfolio, error) {
			return &models.Portfolio{ID: id, UserID: "other-user"}, nil
		},
	}
	svc := &PortfolioService{PortfolioRepo: repo}

	err := svc.Delete("u1", "p1")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
	if repo.deleteCalls != 0 {
		t.Fatalf("expected Delete not to be called")
	}
}

func TestPortfolioService_Delete_Success(t *testing.T) {
	repo := &fakePortfolioRepository{
		getByIDFn: func(id string) (*models.Portfolio, error) {
			return &models.Portfolio{ID: id, UserID: "u1"}, nil
		},
	}
	svc := &PortfolioService{PortfolioRepo: repo}

	err := svc.Delete("u1", "p1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.deleteCalls != 1 || repo.lastDeletedID != "p1" {
		t.Fatalf("expected Delete(p1) called once, got calls=%d id=%q", repo.deleteCalls, repo.lastDeletedID)
	}
}

func TestPortfolioService_GetPortfolioSummary_Unauthorized(t *testing.T) {
	repo := &fakePortfolioRepository{
		getByIDFn: func(id string) (*models.Portfolio, error) {
			return &models.Portfolio{ID: id, UserID: "other", Name: "X"}, nil
		},
	}
	holdings := &fakeHoldingRepository{}
	svc := &PortfolioService{PortfolioRepo: repo, HoldingRepo: holdings}

	_, err := svc.GetPortfolioSummary("u1", "p1")
	if err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestPortfolioService_GetPortfolioSummary_SuccessAndAllocation(t *testing.T) {
	repo := &fakePortfolioRepository{
		getByIDFn: func(id string) (*models.Portfolio, error) {
			return &models.Portfolio{ID: id, UserID: "u1", Name: "Main"}, nil
		},
	}
	holdings := &fakeHoldingRepository{
		getByPortfolioFn: func(portfolioID string) ([]models.Holding, error) {
			return []models.Holding{
				{Symbol: "AAA", AssetType: "stock", Quantity: 10, AvgPrice: 100},
				{Symbol: "BBB", AssetType: "etf", Quantity: 5, AvgPrice: 200},
			}, nil
		},
	}
	svc := &PortfolioService{PortfolioRepo: repo, HoldingRepo: holdings, Prices: market.Passthrough{}}

	s, err := svc.GetPortfolioSummary("u1", "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.PortfolioName != "Main" || s.TotalInvested != 2000 || s.TotalPortfolioValue != 2000 {
		t.Fatalf("unexpected totals: %#v", s)
	}
	if s.TotalProfitLoss != 0 || s.ProfitLossPercentage != 0 {
		t.Fatalf("with passthrough prices P/L should be zero, got %#v", s)
	}
	if len(s.AssetAllocation) != 2 {
		t.Fatalf("expected 2 allocation rows, got %d", len(s.AssetAllocation))
	}
	if s.AssetAllocation[0].Percent != 50 || s.AssetAllocation[1].Percent != 50 {
		t.Fatalf("expected 50/50 allocation, got %#v", s.AssetAllocation)
	}
	if s.AssetAllocation[0].CostBasis != 1000 || s.AssetAllocation[0].CurrentPrice != 100 {
		t.Fatalf("unexpected row 0 cost/price: %#v", s.AssetAllocation[0])
	}
}

func TestPortfolioService_GetPortfolioSummary_MarkToMarket(t *testing.T) {
	repo := &fakePortfolioRepository{
		getByIDFn: func(id string) (*models.Portfolio, error) {
			return &models.Portfolio{ID: id, UserID: "u1", Name: "Main"}, nil
		},
	}
	holdings := &fakeHoldingRepository{
		getByPortfolioFn: func(portfolioID string) ([]models.Holding, error) {
			return []models.Holding{
				{Symbol: "AAA", AssetType: "stock", Quantity: 10, AvgPrice: 100},
			}, nil
		},
	}
	prices := stubUnitPrices{fn: func(symbol string, avg float64) float64 {
		if symbol == "AAA" {
			return 110
		}
		return avg
	}}
	svc := &PortfolioService{PortfolioRepo: repo, HoldingRepo: holdings, Prices: prices}

	s, err := svc.GetPortfolioSummary("u1", "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.TotalInvested != 1000 || s.TotalPortfolioValue != 1100 {
		t.Fatalf("unexpected totals: invested=%f value=%f", s.TotalInvested, s.TotalPortfolioValue)
	}
	if s.TotalProfitLoss != 100 || s.ProfitLossPercentage != 10 {
		t.Fatalf("unexpected P/L: %+v", s)
	}
	if len(s.AssetAllocation) != 1 || s.AssetAllocation[0].Value != 1100 {
		t.Fatalf("unexpected allocation: %#v", s.AssetAllocation)
	}
}

func TestPortfolioService_GetPortfolioSummary_NoHoldingRepo(t *testing.T) {
	repo := &fakePortfolioRepository{
		getByIDFn: func(id string) (*models.Portfolio, error) {
			return &models.Portfolio{ID: id, UserID: "u1"}, nil
		},
	}
	svc := &PortfolioService{PortfolioRepo: repo, HoldingRepo: nil}

	_, err := svc.GetPortfolioSummary("u1", "p1")
	if err == nil || err.Error() != "holding repository not configured" {
		t.Fatalf("expected config error, got %v", err)
	}
}

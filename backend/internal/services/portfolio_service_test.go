package services

import (
	"errors"
	"testing"

	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/repository"
)

type fakePortfolioRepository struct {
	createFn        func(p *models.Portfolio) error
	getByUserFn     func(userID string) ([]models.Portfolio, error)
	getByIDFn       func(id string) (*models.Portfolio, error)
	updateNameFn    func(id string, name string) error
	deleteFn        func(id string) error

	createCalls     int
	lastCreated     *models.Portfolio

	updateNameCalls int
	lastUpdatedID    string
	lastUpdatedName  string

	deleteCalls    int
	lastDeletedID  string
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


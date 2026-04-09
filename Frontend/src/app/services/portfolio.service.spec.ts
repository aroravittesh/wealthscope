import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';

import { environment } from '../../environments/environment';
import { PortfolioService } from './portfolio.service';

describe('PortfolioService', () => {
  let service: PortfolioService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [PortfolioService],
    });

    service = TestBed.inject(PortfolioService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('getHoldings should map backend fields into Holding model', () => {
    const portfolioId = 'pid-1';

    const backendResponse = [
      {
        id: 'h1',
        portfolio_id: portfolioId,
        symbol: 'AAPL',
        asset_type: 'STOCK',
        quantity: '10',
        avg_price: '100.5',
        created_at: '2020-01-01T00:00:00Z',
        updated_at: '2020-01-02T00:00:00Z',
      },
    ];

    service.getHoldings(portfolioId).subscribe(holdings => {
      expect(holdings.length).toBe(1);
      expect(holdings[0].id).toBe('h1');
      expect(holdings[0].portfolioId).toBe(portfolioId);
      expect(holdings[0].symbol).toBe('AAPL');
      expect(holdings[0].assetType).toBe('STOCK');
      expect(holdings[0].quantity).toBe(10);
      expect(holdings[0].avgPrice).toBe(100.5);
      expect(holdings[0].createdAt).toEqual(new Date('2020-01-01T00:00:00Z'));
      expect(holdings[0].updatedAt).toEqual(new Date('2020-01-02T00:00:00Z'));
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/holdings/${portfolioId}`);
    expect(req.request.method).toBe('GET');
    req.flush(backendResponse);
  });

  it('getPortfolioSummary should map backend summary response', () => {
    const portfolioId = 'p-99';
    const backendResponse = {
      portfolio_id: portfolioId,
      portfolio_name: 'Main',
      total_invested: 1000,
      total_portfolio_value: 1100,
      total_profit_loss: 100,
      profit_loss_percentage: 10,
      diversification_score: 87.5,
      volatility_score: 48.25,
      asset_allocation: [
        {
          symbol: 'AAA',
          asset_type: 'stock',
          cost_basis: 1000,
          current_price: 110,
          value: 1100,
          percent: 100,
        },
      ],
    };

    service.getPortfolioSummary(portfolioId).subscribe(s => {
      expect(s.portfolioId).toBe(portfolioId);
      expect(s.portfolioName).toBe('Main');
      expect(s.totalInvested).toBe(1000);
      expect(s.totalPortfolioValue).toBe(1100);
      expect(s.totalProfitLoss).toBe(100);
      expect(s.profitLossPercentage).toBe(10);
      expect(s.diversificationScore).toBe(87.5);
      expect(s.volatilityScore).toBe(48.25);
      expect(s.assetAllocation.length).toBe(1);
      expect(s.assetAllocation[0].symbol).toBe('AAA');
      expect(s.assetAllocation[0].assetType).toBe('stock');
      expect(s.assetAllocation[0].costBasis).toBe(1000);
      expect(s.assetAllocation[0].currentPrice).toBe(110);
      expect(s.assetAllocation[0].value).toBe(1100);
      expect(s.assetAllocation[0].percent).toBe(100);
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/portfolios/${portfolioId}/summary`);
    expect(req.request.method).toBe('GET');
    req.flush(backendResponse);
  });
});


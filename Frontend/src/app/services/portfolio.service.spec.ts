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

  it('getPortfolioSnapshotCompare should map delta payload', () => {
    const portfolioId = 'p-1';
    const backendResponse = {
      portfolio_id: portfolioId,
      from_id: 's-old',
      to_id: 's-new',
      from_at: '2026-04-20T10:00:00.000Z',
      to_at: '2026-04-29T10:00:00.000Z',
      total_value_delta: { absolute: 250, percent: 10 },
      total_invested_delta: { absolute: 100, percent: 5 },
      profit_loss_delta: { absolute: 150, percent: 25 },
      diversification_delta: { absolute: 2.5, percent: 3.2 },
      volatility_delta: { absolute: -1.1, percent: -2.0 },
      allocation_drift: [
        {
          symbol: 'AAPL',
          from_percent: 20,
          to_percent: 30,
          delta_percent: 10,
          from_value: 200,
          to_value: 300,
          delta_value: 100,
        },
      ],
    };

    service.getPortfolioSnapshotCompare(portfolioId, 's-old', 's-new').subscribe(compare => {
      expect(compare.portfolioId).toBe(portfolioId);
      expect(compare.fromId).toBe('s-old');
      expect(compare.toId).toBe('s-new');
      expect(compare.totalValueDelta.absolute).toBe(250);
      expect(compare.profitLossDelta.percent).toBe(25);
      expect(compare.volatilityDelta.absolute).toBe(-1.1);
      expect(compare.allocationDrift.length).toBe(1);
      expect(compare.allocationDrift[0].symbol).toBe('AAPL');
      expect(compare.fromAt.toISOString()).toBe('2026-04-20T10:00:00.000Z');
      expect(compare.toAt.toISOString()).toBe('2026-04-29T10:00:00.000Z');
    });

    const req = httpMock.expectOne(
      `${environment.apiUrl}/portfolios/${portfolioId}/snapshots/compare?from=s-old&to=s-new`
    );
    expect(req.request.method).toBe('GET');
    req.flush(backendResponse);
  });

  it('getPortfolioSnapshotTrend should map trend points', () => {
    const portfolioId = 'p-1';
    const backendResponse = {
      portfolio_id: portfolioId,
      points: [
        {
          snapshot_id: 's-1',
          created_at: '2026-04-20T10:00:00.000Z',
          total_portfolio_value: 1000,
          total_invested: 900,
          total_profit_loss: 100,
          diversification: 50,
          volatility: 40,
        },
      ],
    };

    service.getPortfolioSnapshotTrend(portfolioId, 10).subscribe(trend => {
      expect(trend.portfolioId).toBe(portfolioId);
      expect(trend.points.length).toBe(1);
      expect(trend.points[0].snapshotId).toBe('s-1');
      expect(trend.points[0].totalPortfolioValue).toBe(1000);
      expect(trend.points[0].totalProfitLoss).toBe(100);
      expect(trend.points[0].createdAt.toISOString()).toBe('2026-04-20T10:00:00.000Z');
    });

    const req = httpMock.expectOne(
      `${environment.apiUrl}/portfolios/${portfolioId}/snapshots/trend?limit=10`
    );
    expect(req.request.method).toBe('GET');
    req.flush(backendResponse);
  });
});


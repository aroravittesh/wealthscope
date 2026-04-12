import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { PortfolioService } from './portfolio.service';

describe('PortfolioService', () => {
  let service: PortfolioService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [PortfolioService]
    });
    service = TestBed.inject(PortfolioService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  // 1. getPortfolios
  it('should retrieve mock portfolios (#getPortfolios)', (done) => {
    service.getPortfolios().subscribe(ports => {
      expect(ports.length).toBeGreaterThan(0);
      done();
    });
  });

  // 2. getPortfolioById
  it('should fetch portfolio by id (#getPortfolioById)', () => {
    service.getPortfolioById('1').subscribe();
    const req = httpMock.expectOne({ method: 'GET' });
    expect(req.request.url).toContain('/1');
    req.flush({});
  });

  // 3. createPortfolio
  it('should create new portfolio (#createPortfolio)', () => {
    service.createPortfolio({ name: 'New', description: 'Desc' }).subscribe();
    const req = httpMock.expectOne({ method: 'POST' });
    req.flush({ id: 'abc', name: 'New', description: 'Desc' });
  });

  // 4. updatePortfolio
  it('should update portfolio properly (#updatePortfolio)', () => {
    service.updatePortfolio('99', { name: 'M', description: 'N' }).subscribe();
    const req = httpMock.expectOne({ method: 'PUT' });
    expect(req.request.url).toContain('/99');
    req.flush({});
  });

  // 5. deletePortfolio
  it('should delete specified portfolio (#deletePortfolio)', () => {
    service.deletePortfolio('5').subscribe();
    const req = httpMock.expectOne({ method: 'DELETE' });
    expect(req.request.url).toContain('/5');
    req.flush({});
  });

  // 6. getPortfolioMetrics
  it('should fetch metrics natively or fallback to mock (#getPortfolioMetrics)', () => {
    service.getPortfolioMetrics('1').subscribe(m => {
      expect(m.totalInvested).toBeDefined();
    });
    const req = httpMock.expectOne({ method: 'GET' });
    req.error(new ErrorEvent('Network error')); // Trigger error to hit catchError fallback!
  });

  // 7. getHoldings
  it('should query holdings specifically (#getHoldings)', () => {
    service.getHoldings('p123').subscribe();
    const req = httpMock.expectOne({ method: 'GET' });
    expect(req.request.url).toContain('/p123/holdings');
    req.flush([]);
  });
});

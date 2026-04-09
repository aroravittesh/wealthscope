import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, convertToParamMap, Router } from '@angular/router';
import { of } from 'rxjs';

import { HoldingsComponent } from './holdings.component';
import { PortfolioService } from '../../../services/portfolio.service';
import { Holding, Portfolio } from '../../../models';

describe('HoldingsComponent', () => {
  let fixture: ComponentFixture<HoldingsComponent>;
  let component: HoldingsComponent;

  const portfolioId = '11111111-1111-1111-1111-111111111111';

  const portfolioStub: Portfolio = {
    id: portfolioId,
    userId: 'user-1',
    name: 'My Portfolio',
    description: 'desc',
    totalValue: 0,
    totalInvested: 0,
    totalProfitLoss: 0,
    profitLossPercentage: 0,
    createdAt: new Date(),
    updatedAt: new Date(),
  };

  const holdingsStub: Holding[] = [];

  const routerStub = {
    navigate: jasmine.createSpy('navigate'),
  };

  const portfolioServiceSpy = jasmine.createSpyObj<PortfolioService>(
    'PortfolioService',
    ['getPortfolioById', 'getHoldings', 'addHolding', 'updateHolding', 'deleteHolding']
  );

  beforeEach(async () => {
    portfolioServiceSpy.getPortfolioById.and.returnValue(of(portfolioStub));
    portfolioServiceSpy.getHoldings.and.returnValue(of(holdingsStub));
    portfolioServiceSpy.addHolding.and.returnValue(of({ message: 'holding added' }));
    portfolioServiceSpy.updateHolding.and.returnValue(of({ message: 'holding updated' }));
    portfolioServiceSpy.deleteHolding.and.returnValue(of({ message: 'holding deleted' }));

    await TestBed.configureTestingModule({
      imports: [HoldingsComponent],
      providers: [
        { provide: ActivatedRoute, useValue: { paramMap: of(convertToParamMap({ portfolioId })) } },
        { provide: Router, useValue: routerStub },
        { provide: PortfolioService, useValue: portfolioServiceSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(HoldingsComponent);
    component = fixture.componentInstance;

    fixture.detectChanges();
    await fixture.whenStable();
  });

  it('should uppercase asset_type and symbol when saving a holding', () => {
    component.openAddModal();

    component.holdingForm = {
      symbol: 'aapl',
      assetType: 'stock',
      quantity: 10,
      avgPrice: 100.5,
    };

    component.saveHolding();

    expect(portfolioServiceSpy.addHolding).toHaveBeenCalledWith({
      portfolio_id: portfolioId,
      symbol: 'AAPL',
      asset_type: 'STOCK',
      quantity: 10,
      avg_price: 100.5,
    });

    expect(component.mutating).toBeFalse();
    expect(component.showHoldingModal).toBeFalse();
    expect(component.formErrorMessage).toBeNull();
  });
});


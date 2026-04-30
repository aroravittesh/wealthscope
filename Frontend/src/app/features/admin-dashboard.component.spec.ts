import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';

import { AdminDashboardComponent } from './admin-dashboard.component';
import { AdminService } from '../services/admin.service';
import { AdminAsset, AdminUser } from '../services/admin.service';

describe('AdminDashboardComponent', () => {
  let fixture: ComponentFixture<AdminDashboardComponent>;
  let component: AdminDashboardComponent;
  let adminServiceSpy: jasmine.SpyObj<AdminService>;

  const sampleUser: AdminUser = {
    id: 'u1',
    email: 'user@example.com',
    role: 'USER',
    riskPreference: 'MEDIUM',
  };

  const sampleAsset: AdminAsset = {
    id: 'a1',
    symbol: 'AAPL',
    name: 'Apple Inc.',
    assetType: 'stock',
    currentPrice: 180,
    marketCap: 1000,
    volume: 500,
  };

  const sampleLog = {
    id: 'log-1',
    actorUserId: 'admin-1',
    action: 'ASSET_CREATED',
    entityType: 'asset',
    entityId: 'a1',
    beforeJson: '',
    afterJson: '{"symbol":"AAPL"}',
    createdAt: new Date(),
  };
  const secondLog = {
    id: 'log-2',
    actorUserId: 'admin-2',
    action: 'USER_ROLE_UPDATED',
    entityType: 'user',
    entityId: 'u1',
    beforeJson: '{"role":"USER"}',
    afterJson: '{"role":"ADMIN"}',
    createdAt: new Date('2026-04-30T12:00:00.000Z'),
  };

  beforeEach(async () => {
    adminServiceSpy = jasmine.createSpyObj<AdminService>('AdminService', [
      'getUsers',
      'getAssets',
      'updateUserRole',
      'createAsset',
      'updateAsset',
      'deleteAsset',
      'getAuditLogs',
    ]);

    adminServiceSpy.getUsers.and.returnValue(of([sampleUser]));
    adminServiceSpy.getAssets.and.returnValue(of([sampleAsset]));
    adminServiceSpy.updateUserRole.and.returnValue(of({}));
    adminServiceSpy.createAsset.and.returnValue(of(sampleAsset));
    adminServiceSpy.updateAsset.and.returnValue(of(sampleAsset));
    adminServiceSpy.deleteAsset.and.returnValue(of({}));
    adminServiceSpy.getAuditLogs.and.returnValue(of([sampleLog, secondLog]));

    await TestBed.configureTestingModule({
      imports: [AdminDashboardComponent],
      providers: [{ provide: AdminService, useValue: adminServiceSpy }],
    }).compileComponents();

    fixture = TestBed.createComponent(AdminDashboardComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
    await fixture.whenStable();
  });

  it('should load users and assets on init', () => {
    expect(adminServiceSpy.getUsers).toHaveBeenCalled();
    expect(adminServiceSpy.getAssets).toHaveBeenCalled();
    expect(adminServiceSpy.getAuditLogs).toHaveBeenCalled();
    expect(component.users.length).toBe(1);
    expect(component.assets.length).toBe(1);
    expect(component.auditLogs.length).toBe(2);
  });

  it('reloadAll should clear messages and reload data', () => {
    component.errorMessage = 'previous error';
    component.successMessage = 'previous ok';
    component.reloadAll();

    expect(component.errorMessage).toBeNull();
    expect(component.successMessage).toBeNull();
    expect(adminServiceSpy.getUsers.calls.count()).toBeGreaterThan(1);
    expect(adminServiceSpy.getAssets.calls.count()).toBeGreaterThan(1);
  });

  it('saveUserRole should set success message on success', () => {
    component.saveUserRole(sampleUser);

    expect(adminServiceSpy.updateUserRole).toHaveBeenCalledWith('u1', 'USER');
    expect(component.successMessage).toContain('user@example.com');
    expect(component.updatingUserId).toBeNull();
  });

  it('saveUserRole should set error message on failure', () => {
    adminServiceSpy.updateUserRole.and.returnValue(throwError(() => ({ error: 'nope' })));

    component.saveUserRole(sampleUser);

    expect(component.errorMessage).toBe('nope');
    expect(component.updatingUserId).toBeNull();
  });

  it('submitAssetForm should create when not editing', () => {
    component.activeTab = 'assets';
    component.editAssetId = null;
    const expectedPayload = {
      symbol: 'MSFT',
      name: 'Microsoft',
      assetType: 'stock',
      currentPrice: 300,
      marketCap: 2000,
      volume: 100,
    };
    component.assetForm = { ...expectedPayload };

    component.submitAssetForm();

    expect(adminServiceSpy.createAsset).toHaveBeenCalledWith(expectedPayload);
    expect(adminServiceSpy.updateAsset).not.toHaveBeenCalled();
    expect(component.savingAsset).toBeFalse();
    expect(component.successMessage).toContain('created');
  });

  it('submitAssetForm should update when editAssetId is set', () => {
    component.activeTab = 'assets';
    component.editAssetId = 'a1';
    const expectedPayload = {
      symbol: 'AAPL',
      name: 'Apple Inc.',
      assetType: 'stock',
      currentPrice: 181,
      marketCap: 1001,
      volume: 501,
    };
    component.assetForm = { ...expectedPayload };

    component.submitAssetForm();

    expect(adminServiceSpy.updateAsset).toHaveBeenCalledWith('a1', expectedPayload);
    expect(component.successMessage).toContain('updated');
  });

  it('startEditAsset should populate the form', () => {
    component.startEditAsset(sampleAsset);

    expect(component.editAssetId).toBe('a1');
    expect(component.assetForm.symbol).toBe('AAPL');
    expect(component.assetForm.currentPrice).toBe(180);
  });

  it('removeAsset should refresh assets on success', () => {
    const initialCalls = adminServiceSpy.getAssets.calls.count();

    component.removeAsset(sampleAsset);

    expect(adminServiceSpy.deleteAsset).toHaveBeenCalledWith('a1');
    expect(component.deletingAssetId).toBeNull();
    expect(component.successMessage).toContain('AAPL');
    expect(adminServiceSpy.getAssets.calls.count()).toBeGreaterThan(initialCalls);
  });

  it('resetAssetForm should clear edit state and form', () => {
    component.editAssetId = 'a1';
    component.assetForm = {
      symbol: 'X',
      name: 'Y',
      assetType: 'stock',
      currentPrice: 1,
      marketCap: 2,
      volume: 3,
    };

    component.resetAssetForm();

    expect(component.editAssetId).toBeNull();
    expect(component.assetForm.symbol).toBe('');
    expect(component.assetForm.assetType).toBe('stock');
  });

  it('filteredAuditLogs should apply actor and action filters', () => {
    component.activityUserFilter = 'admin-2';
    component.activityActionFilter = 'USER_ROLE_UPDATED';

    expect(component.filteredAuditLogs.length).toBe(1);
    expect(component.filteredAuditLogs[0].id).toBe('log-2');
  });

  it('filteredAuditLogs should apply search term and date range', () => {
    component.activitySearchTerm = 'aapl';
    component.activityStartDate = '2026-04-01';
    component.activityEndDate = '2026-04-29';

    const results = component.filteredAuditLogs;
    expect(results.length).toBe(1);
    expect(results[0].id).toBe('log-1');
  });

  it('clearActivityFilters should reset all activity filter fields', () => {
    component.activitySearchTerm = 'asset';
    component.activityUserFilter = 'admin-1';
    component.activityActionFilter = 'ASSET_CREATED';
    component.activityStartDate = '2026-04-01';
    component.activityEndDate = '2026-04-30';

    component.clearActivityFilters();

    expect(component.activitySearchTerm).toBe('');
    expect(component.activityUserFilter).toBe('');
    expect(component.activityActionFilter).toBe('');
    expect(component.activityStartDate).toBe('');
    expect(component.activityEndDate).toBe('');
  });
});

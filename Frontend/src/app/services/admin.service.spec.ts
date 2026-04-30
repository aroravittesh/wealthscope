import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';

import { environment } from '../../environments/environment';
import { AdminService, AdminAssetPayload } from './admin.service';

describe('AdminService', () => {
  let service: AdminService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [AdminService],
    });

    service = TestBed.inject(AdminService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('getUsers should map backend rows to AdminUser', () => {
    const backend = [
      {
        id: 1,
        email: 'a@b.com',
        role: 'ADMIN',
        risk_preference: 'LOW',
      },
    ];

    service.getUsers().subscribe(users => {
      expect(users.length).toBe(1);
      expect(users[0].id).toBe('1');
      expect(users[0].email).toBe('a@b.com');
      expect(users[0].role).toBe('ADMIN');
      expect(users[0].riskPreference).toBe('LOW');
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/admin/users`);
    expect(req.request.method).toBe('GET');
    req.flush(backend);
  });

  it('updateUserRole should PATCH role endpoint', () => {
    service.updateUserRole('u1', 'USER').subscribe(res => {
      expect(res.message).toBe('ok');
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/admin/users/u1/role`);
    expect(req.request.method).toBe('PATCH');
    expect(req.request.body).toEqual({ role: 'USER' });
    req.flush({ message: 'ok' });
  });

  it('getAssets should map backend rows to AdminAsset', () => {
    const backend = [
      {
        id: 10,
        symbol: 'AAPL',
        name: 'Apple',
        asset_type: 'stock',
        current_price: '150.25',
        market_cap: '1000',
        volume: '500',
      },
    ];

    service.getAssets().subscribe(assets => {
      expect(assets.length).toBe(1);
      expect(assets[0].id).toBe('10');
      expect(assets[0].symbol).toBe('AAPL');
      expect(assets[0].assetType).toBe('stock');
      expect(assets[0].currentPrice).toBe(150.25);
      expect(assets[0].marketCap).toBe(1000);
      expect(assets[0].volume).toBe(500);
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/admin/assets`);
    expect(req.request.method).toBe('GET');
    req.flush(backend);
  });

  it('createAsset should POST mapped payload and map response', () => {
    const payload: AdminAssetPayload = {
      symbol: 'MSFT',
      name: 'Microsoft',
      assetType: 'stock',
      currentPrice: 300,
      marketCap: 2000,
      volume: 100,
    };

    service.createAsset(payload).subscribe(asset => {
      expect(asset.id).toBe('99');
      expect(asset.symbol).toBe('MSFT');
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/admin/assets`);
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({
      symbol: 'MSFT',
      name: 'Microsoft',
      asset_type: 'stock',
      current_price: 300,
      market_cap: 2000,
      volume: 100,
    });
    req.flush({
      id: 99,
      symbol: 'MSFT',
      name: 'Microsoft',
      asset_type: 'stock',
      current_price: 300,
      market_cap: 2000,
      volume: 100,
    });
  });

  it('updateAsset should PUT mapped payload', () => {
    const payload: AdminAssetPayload = {
      symbol: 'MSFT',
      name: 'Microsoft',
      assetType: 'stock',
      currentPrice: 301,
      marketCap: 2001,
      volume: 101,
    };

    service.updateAsset('asset-1', payload).subscribe(asset => {
      expect(asset.id).toBe('asset-1');
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/admin/assets/asset-1`);
    expect(req.request.method).toBe('PUT');
    req.flush({
      id: 'asset-1',
      symbol: 'MSFT',
      name: 'Microsoft',
      asset_type: 'stock',
      current_price: 301,
      market_cap: 2001,
      volume: 101,
    });
  });

  it('deleteAsset should DELETE asset by id', () => {
    service.deleteAsset('del-1').subscribe(res => {
      expect(res.message).toBe('gone');
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/admin/assets/del-1`);
    expect(req.request.method).toBe('DELETE');
    req.flush({ message: 'gone' });
  });

  it('getAuditLogs should map audit log rows', () => {
    const createdAt = '2026-04-29T10:00:00.000Z';
    const backend = [
      {
        id: 'log-1',
        actor_user_id: 'admin-1',
        action: 'ASSET_CREATED',
        entity_type: 'asset',
        entity_id: 'asset-1',
        before_json: '',
        after_json: '{"symbol":"AAPL"}',
        created_at: createdAt,
      },
    ];

    service.getAuditLogs(25).subscribe(logs => {
      expect(logs.length).toBe(1);
      expect(logs[0].id).toBe('log-1');
      expect(logs[0].actorUserId).toBe('admin-1');
      expect(logs[0].action).toBe('ASSET_CREATED');
      expect(logs[0].entityType).toBe('asset');
      expect(logs[0].entityId).toBe('asset-1');
      expect(logs[0].afterJson).toContain('AAPL');
      expect(logs[0].createdAt.toISOString()).toBe(createdAt);
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/admin/audit-logs?limit=25`);
    expect(req.request.method).toBe('GET');
    req.flush(backend);
  });
});

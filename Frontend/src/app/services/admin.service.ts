import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { environment } from '../../environments/environment';

export interface AdminUser {
  id: string;
  email: string;
  role: 'USER' | 'ADMIN';
  riskPreference?: string;
}

export interface AdminAsset {
  id: string;
  symbol: string;
  name: string;
  assetType: string;
  currentPrice: number;
  marketCap: number;
  volume: number;
}

export interface AdminAssetPayload {
  symbol: string;
  name: string;
  assetType: string;
  currentPrice: number;
  marketCap: number;
  volume: number;
}

@Injectable({
  providedIn: 'root'
})
export class AdminService {
  private adminApiUrl = `${environment.apiUrl}/admin`;

  constructor(private http: HttpClient) {}

  getUsers(): Observable<AdminUser[]> {
    return this.http
      .get<any[]>(`${this.adminApiUrl}/users`)
      .pipe(map(rows => (rows || []).map(row => this.mapAdminUser(row))));
  }

  updateUserRole(userId: string, role: 'USER' | 'ADMIN'): Observable<{ message?: string }> {
    return this.http.patch<{ message?: string }>(
      `${this.adminApiUrl}/users/${userId}/role`,
      { role }
    );
  }

  getAssets(): Observable<AdminAsset[]> {
    return this.http
      .get<any[]>(`${this.adminApiUrl}/assets`)
      .pipe(map(rows => (rows || []).map(row => this.mapAdminAsset(row))));
  }

  createAsset(payload: AdminAssetPayload): Observable<AdminAsset> {
    return this.http
      .post<any>(`${this.adminApiUrl}/assets`, this.mapAssetPayload(payload))
      .pipe(map(row => this.mapAdminAsset(row)));
  }

  updateAsset(assetId: string, payload: AdminAssetPayload): Observable<AdminAsset> {
    return this.http
      .put<any>(`${this.adminApiUrl}/assets/${assetId}`, this.mapAssetPayload(payload))
      .pipe(map(row => this.mapAdminAsset(row)));
  }

  deleteAsset(assetId: string): Observable<{ message?: string }> {
    return this.http.delete<{ message?: string }>(`${this.adminApiUrl}/assets/${assetId}`);
  }

  private mapAdminUser(row: any): AdminUser {
    return {
      id: String(row.id),
      email: row.email,
      role: (row.role ?? 'USER') as 'USER' | 'ADMIN',
      riskPreference: row.risk_preference ?? row.riskPreference
    };
  }

  private mapAdminAsset(row: any): AdminAsset {
    return {
      id: String(row.id),
      symbol: row.symbol ?? '',
      name: row.name ?? '',
      assetType: row.asset_type ?? row.assetType ?? 'stock',
      currentPrice: Number(row.current_price ?? row.currentPrice ?? 0),
      marketCap: Number(row.market_cap ?? row.marketCap ?? 0),
      volume: Number(row.volume ?? 0)
    };
  }

  private mapAssetPayload(payload: AdminAssetPayload): any {
    return {
      symbol: payload.symbol,
      name: payload.name,
      asset_type: payload.assetType,
      current_price: payload.currentPrice,
      market_cap: payload.marketCap,
      volume: payload.volume
    };
  }
}

import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AdminAsset, AdminAssetPayload, AdminService, AdminUser } from '../services/admin.service';

@Component({
  selector: 'app-admin-dashboard',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      <div class="bg-slate-800/60 rounded-xl p-8 border border-slate-700 shadow-lg">
        <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-6">
          <div>
            <h2 class="text-3xl font-bold text-yellow-300">Admin Dashboard</h2>
            <p class="text-slate-400 text-sm mt-1">Manage users and assets with backend admin APIs.</p>
          </div>
          <button
            type="button"
            (click)="reloadAll()"
            class="bg-slate-700 hover:bg-slate-600 text-white px-4 py-2 rounded-lg text-sm font-medium w-fit"
            [disabled]="loadingUsers || loadingAssets"
          >
            Refresh
          </button>
        </div>

        <div *ngIf="errorMessage" class="mb-4 bg-red-900/30 border border-red-500/40 text-red-200 rounded-lg p-3 text-sm">
          {{ errorMessage }}
        </div>
        <div *ngIf="successMessage" class="mb-4 bg-emerald-900/30 border border-emerald-500/40 text-emerald-200 rounded-lg p-3 text-sm">
          {{ successMessage }}
        </div>

        <div class="flex gap-2 mb-6">
          <button
            type="button"
            (click)="activeTab = 'users'"
            class="px-4 py-2 rounded-lg text-sm font-semibold border"
            [ngClass]="activeTab === 'users' ? 'bg-yellow-600 text-white border-yellow-500' : 'bg-slate-900 text-slate-300 border-slate-700'"
          >
            Users
          </button>
          <button
            type="button"
            (click)="activeTab = 'assets'"
            class="px-4 py-2 rounded-lg text-sm font-semibold border"
            [ngClass]="activeTab === 'assets' ? 'bg-yellow-600 text-white border-yellow-500' : 'bg-slate-900 text-slate-300 border-slate-700'"
          >
            Assets
          </button>
        </div>

        <section *ngIf="activeTab === 'users'" class="bg-slate-900 rounded-xl border border-slate-700 overflow-hidden">
          <div class="p-4 border-b border-slate-700">
            <h3 class="text-lg font-semibold text-white">User Management</h3>
            <p class="text-slate-400 text-xs mt-1">Update account role using admin role endpoint.</p>
          </div>

          <div *ngIf="loadingUsers" class="p-4 text-slate-300 text-sm">Loading users...</div>
          <div *ngIf="!loadingUsers && users.length === 0" class="p-4 text-slate-400 text-sm">No users found.</div>

          <div *ngIf="!loadingUsers && users.length > 0" class="overflow-x-auto">
            <table class="min-w-full text-sm">
              <thead class="bg-slate-800 text-slate-300">
                <tr>
                  <th class="text-left p-3">Email</th>
                  <th class="text-left p-3">Risk</th>
                  <th class="text-left p-3">Role</th>
                  <th class="text-left p-3">Action</th>
                </tr>
              </thead>
              <tbody>
                <tr *ngFor="let user of users" class="border-t border-slate-800">
                  <td class="p-3 text-slate-100">{{ user.email }}</td>
                  <td class="p-3 text-slate-400">{{ user.riskPreference || '-' }}</td>
                  <td class="p-3">
                    <select
                      [(ngModel)]="user.role"
                      [name]="'role-' + user.id"
                      class="bg-slate-800 text-slate-200 border border-slate-600 rounded px-2 py-1"
                    >
                      <option value="USER">USER</option>
                      <option value="ADMIN">ADMIN</option>
                    </select>
                  </td>
                  <td class="p-3">
                    <button
                      type="button"
                      (click)="saveUserRole(user)"
                      class="bg-blue-600 hover:bg-blue-500 text-white px-3 py-1 rounded text-xs font-semibold"
                      [disabled]="updatingUserId === user.id"
                    >
                      {{ updatingUserId === user.id ? 'Saving...' : 'Save Role' }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <section *ngIf="activeTab === 'assets'" class="space-y-6">
          <div class="bg-slate-900 rounded-xl border border-slate-700 p-4">
            <h3 class="text-lg font-semibold text-white mb-3">
              {{ editAssetId ? 'Edit Asset' : 'Create Asset' }}
            </h3>
            <form (ngSubmit)="submitAssetForm()" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              <input
                type="text"
                placeholder="Symbol (e.g. AAPL)"
                [(ngModel)]="assetForm.symbol"
                name="symbol"
                class="bg-slate-800 border border-slate-600 rounded px-3 py-2 text-slate-200"
                required
              />
              <input
                type="text"
                placeholder="Name"
                [(ngModel)]="assetForm.name"
                name="name"
                class="bg-slate-800 border border-slate-600 rounded px-3 py-2 text-slate-200"
                required
              />
              <input
                type="text"
                placeholder="Asset Type (stock/crypto/etf)"
                [(ngModel)]="assetForm.assetType"
                name="assetType"
                class="bg-slate-800 border border-slate-600 rounded px-3 py-2 text-slate-200"
                required
              />
              <input
                type="number"
                min="0"
                step="0.01"
                placeholder="Current Price"
                [(ngModel)]="assetForm.currentPrice"
                name="currentPrice"
                class="bg-slate-800 border border-slate-600 rounded px-3 py-2 text-slate-200"
                required
              />
              <input
                type="number"
                min="0"
                step="1"
                placeholder="Market Cap"
                [(ngModel)]="assetForm.marketCap"
                name="marketCap"
                class="bg-slate-800 border border-slate-600 rounded px-3 py-2 text-slate-200"
                required
              />
              <input
                type="number"
                min="0"
                step="1"
                placeholder="Volume"
                [(ngModel)]="assetForm.volume"
                name="volume"
                class="bg-slate-800 border border-slate-600 rounded px-3 py-2 text-slate-200"
                required
              />

              <div class="md:col-span-2 lg:col-span-3 flex gap-2">
                <button
                  type="submit"
                  class="bg-yellow-600 hover:bg-yellow-500 text-white px-4 py-2 rounded font-semibold"
                  [disabled]="savingAsset"
                >
                  {{ savingAsset ? 'Saving...' : (editAssetId ? 'Update Asset' : 'Create Asset') }}
                </button>
                <button
                  *ngIf="editAssetId"
                  type="button"
                  (click)="resetAssetForm()"
                  class="bg-slate-700 hover:bg-slate-600 text-white px-4 py-2 rounded font-semibold"
                >
                  Cancel Edit
                </button>
              </div>
            </form>
          </div>

          <div class="bg-slate-900 rounded-xl border border-slate-700 overflow-hidden">
            <div class="p-4 border-b border-slate-700">
              <h3 class="text-lg font-semibold text-white">Asset Management</h3>
            </div>

            <div *ngIf="loadingAssets" class="p-4 text-slate-300 text-sm">Loading assets...</div>
            <div *ngIf="!loadingAssets && assets.length === 0" class="p-4 text-slate-400 text-sm">No assets found.</div>

            <div *ngIf="!loadingAssets && assets.length > 0" class="overflow-x-auto">
              <table class="min-w-full text-sm">
                <thead class="bg-slate-800 text-slate-300">
                  <tr>
                    <th class="text-left p-3">Symbol</th>
                    <th class="text-left p-3">Name</th>
                    <th class="text-left p-3">Type</th>
                    <th class="text-left p-3">Price</th>
                    <th class="text-left p-3">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  <tr *ngFor="let asset of assets" class="border-t border-slate-800">
                    <td class="p-3 text-slate-100">{{ asset.symbol }}</td>
                    <td class="p-3 text-slate-200">{{ asset.name }}</td>
                    <td class="p-3 text-slate-400 uppercase">{{ asset.assetType }}</td>
                    <td class="p-3 text-slate-200">\${{ asset.currentPrice.toFixed(2) }}</td>
                    <td class="p-3 flex gap-2">
                      <button
                        type="button"
                        (click)="startEditAsset(asset)"
                        class="bg-blue-600 hover:bg-blue-500 text-white px-3 py-1 rounded text-xs font-semibold"
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        (click)="removeAsset(asset)"
                        class="bg-red-600 hover:bg-red-500 text-white px-3 py-1 rounded text-xs font-semibold"
                        [disabled]="deletingAssetId === asset.id"
                      >
                        {{ deletingAssetId === asset.id ? 'Deleting...' : 'Delete' }}
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>
      </div>
    </div>
  `,
  styles: []
})
export class AdminDashboardComponent implements OnInit {
  activeTab: 'users' | 'assets' = 'users';

  users: AdminUser[] = [];
  assets: AdminAsset[] = [];

  loadingUsers = false;
  loadingAssets = false;
  updatingUserId: string | null = null;
  savingAsset = false;
  deletingAssetId: string | null = null;

  errorMessage: string | null = null;
  successMessage: string | null = null;

  editAssetId: string | null = null;
  assetForm: AdminAssetPayload = this.emptyAssetForm();

  constructor(private adminService: AdminService) {}

  ngOnInit(): void {
    this.reloadAll();
  }

  reloadAll(): void {
    this.errorMessage = null;
    this.successMessage = null;
    this.loadUsers();
    this.loadAssets();
  }

  saveUserRole(user: AdminUser): void {
    this.errorMessage = null;
    this.successMessage = null;
    this.updatingUserId = user.id;

    this.adminService.updateUserRole(user.id, user.role).subscribe({
      next: () => {
        this.updatingUserId = null;
        this.successMessage = `Updated role for ${user.email}.`;
      },
      error: err => {
        this.updatingUserId = null;
        this.errorMessage = err?.error?.message ?? err?.error ?? 'Failed to update user role.';
      }
    });
  }

  submitAssetForm(): void {
    this.errorMessage = null;
    this.successMessage = null;
    this.savingAsset = true;

    const request$ = this.editAssetId
      ? this.adminService.updateAsset(this.editAssetId, this.assetForm)
      : this.adminService.createAsset(this.assetForm);

    request$.subscribe({
      next: () => {
        this.savingAsset = false;
        this.successMessage = this.editAssetId ? 'Asset updated successfully.' : 'Asset created successfully.';
        this.resetAssetForm();
        this.loadAssets();
      },
      error: err => {
        this.savingAsset = false;
        this.errorMessage = err?.error?.message ?? err?.error ?? 'Failed to save asset.';
      }
    });
  }

  startEditAsset(asset: AdminAsset): void {
    this.editAssetId = asset.id;
    this.assetForm = {
      symbol: asset.symbol,
      name: asset.name,
      assetType: asset.assetType,
      currentPrice: asset.currentPrice,
      marketCap: asset.marketCap,
      volume: asset.volume
    };
  }

  removeAsset(asset: AdminAsset): void {
    this.errorMessage = null;
    this.successMessage = null;
    this.deletingAssetId = asset.id;

    this.adminService.deleteAsset(asset.id).subscribe({
      next: () => {
        this.deletingAssetId = null;
        this.successMessage = `Deleted ${asset.symbol}.`;
        this.loadAssets();
      },
      error: err => {
        this.deletingAssetId = null;
        this.errorMessage = err?.error?.message ?? err?.error ?? 'Failed to delete asset.';
      }
    });
  }

  resetAssetForm(): void {
    this.editAssetId = null;
    this.assetForm = this.emptyAssetForm();
  }

  private loadUsers(): void {
    this.loadingUsers = true;
    this.adminService.getUsers().subscribe({
      next: users => {
        this.loadingUsers = false;
        this.users = users;
      },
      error: err => {
        this.loadingUsers = false;
        this.errorMessage = err?.error?.message ?? err?.error ?? 'Failed to load users.';
      }
    });
  }

  private loadAssets(): void {
    this.loadingAssets = true;
    this.adminService.getAssets().subscribe({
      next: assets => {
        this.loadingAssets = false;
        this.assets = assets;
      },
      error: err => {
        this.loadingAssets = false;
        this.errorMessage = err?.error?.message ?? err?.error ?? 'Failed to load assets.';
      }
    });
  }

  private emptyAssetForm(): AdminAssetPayload {
    return {
      symbol: '',
      name: '',
      assetType: 'stock',
      currentPrice: 0,
      marketCap: 0,
      volume: 0
    };
  }
}

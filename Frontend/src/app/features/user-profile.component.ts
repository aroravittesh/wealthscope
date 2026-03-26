import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AuthService } from '../services/auth.service';

@Component({
  selector: 'app-user-profile',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="max-w-lg mx-auto bg-slate-800 bg-opacity-60 rounded-xl p-8 mt-10 border border-slate-700 shadow-lg">
      <h2 class="text-2xl font-bold text-white mb-6">User Profile</h2>

      <div *ngIf="errorMessage" class="mb-4 text-red-400 text-sm">
        {{ errorMessage }}
      </div>
      <div *ngIf="successMessage" class="mb-4 text-emerald-400 text-sm">
        {{ successMessage }}
      </div>

      <!-- Profile section -->
      <form (ngSubmit)="onSaveProfile()" class="space-y-4 mb-8">
        <div>
          <label class="block text-slate-300 mb-1">Email</label>
          <input
            class="w-full px-4 py-2 rounded bg-slate-700 border border-slate-600 text-slate-300 cursor-not-allowed"
            [value]="email"
            disabled
          />
        </div>

        <div>
          <label class="block text-slate-300 mb-1">Risk Preference</label>
          <select
            class="w-full px-4 py-2 rounded bg-slate-700 border border-slate-600 text-white"
            [(ngModel)]="riskPreference"
            name="riskPreference"
          >
            <option value="LOW">Low</option>
            <option value="MEDIUM">Medium</option>
            <option value="HIGH">High</option>
          </select>
        </div>

        <button
          type="submit"
          class="w-full bg-blue-600 hover:bg-blue-500 text-white py-2 rounded mt-4 disabled:opacity-60"
          [disabled]="isSavingProfile"
        >
          {{ isSavingProfile ? 'Saving...' : 'Update Profile' }}
        </button>
      </form>

      <!-- Change password section -->
      <div class="border-t border-slate-700 pt-6">
        <h3 class="text-lg font-semibold text-white mb-4">Change Password</h3>

        <form (ngSubmit)="onChangePassword()" class="space-y-4">
          <div>
            <label class="block text-slate-300 mb-1">Current Password</label>
            <input
              type="password"
              class="w-full px-4 py-2 rounded bg-slate-700 border border-slate-600 text-white"
              [(ngModel)]="oldPassword"
              name="oldPassword"
              required
            />
          </div>

          <div>
            <label class="block text-slate-300 mb-1">New Password</label>
            <input
              type="password"
              class="w-full px-4 py-2 rounded bg-slate-700 border border-slate-600 text-white"
              [(ngModel)]="newPassword"
              name="newPassword"
              required
              minlength="6"
            />
          </div>

          <div>
            <label class="block text-slate-300 mb-1">Confirm New Password</label>
            <input
              type="password"
              class="w-full px-4 py-2 rounded bg-slate-700 border border-slate-600 text-white"
              [(ngModel)]="confirmNewPassword"
              name="confirmNewPassword"
              required
              minlength="6"
            />
          </div>

          <button
            type="submit"
            class="w-full bg-indigo-600 hover:bg-indigo-500 text-white py-2 rounded mt-4 disabled:opacity-60"
            [disabled]="isChangingPassword"
          >
            {{ isChangingPassword ? 'Updating...' : 'Change Password' }}
          </button>
        </form>
      </div>
    </div>
  `
})
export class UserProfileComponent implements OnInit {
  email = '';
  riskPreference: 'LOW' | 'MEDIUM' | 'HIGH' = 'MEDIUM';

  oldPassword = '';
  newPassword = '';
  confirmNewPassword = '';

  isSavingProfile = false;
  isChangingPassword = false;
  errorMessage: string | null = null;
  successMessage: string | null = null;

  constructor(private authService: AuthService) {}

  ngOnInit(): void {
    this.loadProfile();
  }

  private loadProfile(): void {
    this.authService.getProfile().subscribe({
      next: (profile) => {
        this.email = profile.email;
        const risk = profile.risk_preference?.toUpperCase() as 'LOW' | 'MEDIUM' | 'HIGH';
        if (risk === 'LOW' || risk === 'MEDIUM' || risk === 'HIGH') {
          this.riskPreference = risk;
        }
      },
      error: () => {
        this.errorMessage = 'Failed to load profile.';
      }
    });
  }

  onSaveProfile(): void {
    this.isSavingProfile = true;
    this.errorMessage = null;
    this.successMessage = null;

    this.authService.updateRiskPreference(this.riskPreference).subscribe({
      next: () => {
        this.isSavingProfile = false;
        this.successMessage = 'Profile updated successfully.';
      },
      error: () => {
        this.isSavingProfile = false;
        this.errorMessage = 'Failed to update profile.';
      }
    });
  }

  onChangePassword(): void {
    this.errorMessage = null;
    this.successMessage = null;

    if (!this.oldPassword || !this.newPassword || !this.confirmNewPassword) {
      this.errorMessage = 'Please fill in all password fields.';
      return;
    }

    if (this.newPassword !== this.confirmNewPassword) {
      this.errorMessage = 'New passwords do not match.';
      return;
    }

    this.isChangingPassword = true;

    this.authService.changePassword(this.oldPassword, this.newPassword).subscribe({
      next: () => {
        this.isChangingPassword = false;
        this.successMessage = 'Password changed successfully.';
        this.oldPassword = '';
        this.newPassword = '';
        this.confirmNewPassword = '';
      },
      error: (err) => {
        this.isChangingPassword = false;
        this.errorMessage = err?.error || 'Failed to change password.';
      }
    });
  }
}

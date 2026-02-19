import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-admin-dashboard',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="max-w-5xl mx-auto bg-slate-800 bg-opacity-60 rounded-xl p-8 mt-10 border border-slate-700 shadow-lg">
      <h2 class="text-3xl font-bold text-yellow-300 mb-6">Admin Dashboard</h2>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div class="bg-slate-900 rounded-lg p-6 border border-yellow-700">
          <h3 class="text-lg font-semibold text-yellow-200 mb-2">User Management</h3>
          <p class="text-slate-400 mb-4">View, edit, and manage users.</p>
          <button class="bg-yellow-600 text-white px-4 py-2 rounded">Go to Users</button>
        </div>
        <div class="bg-slate-900 rounded-lg p-6 border border-yellow-700">
          <h3 class="text-lg font-semibold text-yellow-200 mb-2">Asset Management</h3>
          <p class="text-slate-400 mb-4">Manage asset master data.</p>
          <button class="bg-yellow-600 text-white px-4 py-2 rounded">Go to Assets</button>
        </div>
        <div class="bg-slate-900 rounded-lg p-6 border border-yellow-700">
          <h3 class="text-lg font-semibold text-yellow-200 mb-2">System Operations</h3>
          <p class="text-slate-400 mb-4">Perform system-level actions.</p>
          <button class="bg-yellow-600 text-white px-4 py-2 rounded">System Ops</button>
        </div>
      </div>
      <div class="bg-slate-900 rounded-lg p-6 border border-yellow-700 mt-8">
        <h3 class="text-lg font-semibold text-yellow-200 mb-2">System Metrics (Stub)</h3>
        <ul class="text-slate-400 text-sm list-disc ml-6">
          <li>Active users: 123</li>
          <li>Assets tracked: 45</li>
          <li>API uptime: 99.99%</li>
          <li>Last backup: 2026-02-13 23:00 UTC</li>
        </ul>
      </div>
    </div>
  `,
  styles: []
})
export class AdminDashboardComponent {}

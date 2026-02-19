import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-system-health',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="max-w-xl mx-auto bg-slate-800 bg-opacity-60 rounded-xl p-8 mt-10 border border-slate-700 shadow-lg">
      <h2 class="text-2xl font-bold text-green-300 mb-6">System Health & Observability (Stub)</h2>
      <div class="bg-slate-900 rounded-lg p-6 border border-green-700 mt-4">
        <h3 class="text-lg font-semibold text-green-200 mb-2">Health Metrics</h3>
        <ul class="text-slate-400 text-sm list-disc ml-6">
          <li>API Status: <span class="text-green-400">Healthy</span></li>
          <li>Database Status: <span class="text-green-400">Connected</span></li>
          <li>Uptime: 99.99%</li>
          <li>Last Health Check: 2026-02-14 09:00 UTC</li>
        </ul>
      </div>
      <div class="bg-slate-900 rounded-lg p-6 border border-green-700 mt-4">
        <h3 class="text-lg font-semibold text-green-200 mb-2">Monitoring Hooks (Stub)</h3>
        <ul class="text-slate-400 text-sm list-disc ml-6">
          <li>Prometheus metrics endpoint: <code>/metrics</code></li>
          <li>Health check API: <code>/health</code></li>
          <li>Log stream: <code>/logs</code></li>
        </ul>
      </div>
    </div>
  `,
  styles: []
})
export class SystemHealthComponent {}

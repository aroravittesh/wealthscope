import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-api',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-purple-950">
      <div class="container mx-auto px-6 py-12">
        <h1 class="text-4xl font-bold text-white mb-8">API Documentation</h1>
        <div class="space-y-8">
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Base URL</h3>
            <p class="text-slate-300 font-mono">https://api.finsight.com/v1</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Authentication</h3>
            <p class="text-slate-300 mb-4">All API requests require an API key in the Authorization header:</p>
            <p class="text-slate-300 font-mono">Authorization: Bearer YOUR_API_KEY</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Endpoints</h3>
            <ul class="space-y-3 text-slate-300">
              <li><span class="font-mono bg-slate-900 px-2 py-1 rounded">GET /portfolios</span> - List user portfolios</li>
              <li><span class="font-mono bg-slate-900 px-2 py-1 rounded">POST /transactions</span> - Create transaction</li>
              <li><span class="font-mono bg-slate-900 px-2 py-1 rounded">GET /analytics</span> - Get analytics data</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  `
})
export class ApiComponent {}

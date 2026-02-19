import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-docs',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-purple-950">
      <div class="container mx-auto px-6 py-12">
        <h1 class="text-4xl font-bold text-white mb-8">Documentation</h1>
        <div class="grid md:grid-cols-2 gap-8">
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Getting Started</h3>
            <ul class="text-slate-300 space-y-2">
              <li><a href="#" class="hover:text-blue-300 transition">Installation Guide</a></li>
              <li><a href="#" class="hover:text-blue-300 transition">Quick Start</a></li>
              <li><a href="#" class="hover:text-blue-300 transition">Basic Concepts</a></li>
            </ul>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">API Reference</h3>
            <ul class="text-slate-300 space-y-2">
              <li><a href="#" class="hover:text-blue-300 transition">REST API</a></li>
              <li><a href="#" class="hover:text-blue-300 transition">Authentication</a></li>
              <li><a href="#" class="hover:text-blue-300 transition">Endpoints</a></li>
            </ul>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Guides</h3>
            <ul class="text-slate-300 space-y-2">
              <li><a href="#" class="hover:text-blue-300 transition">Portfolio Setup</a></li>
              <li><a href="#" class="hover:text-blue-300 transition">Analytics Dashboard</a></li>
              <li><a href="#" class="hover:text-blue-300 transition">Reports</a></li>
            </ul>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Troubleshooting</h3>
            <ul class="text-slate-300 space-y-2">
              <li><a href="#" class="hover:text-blue-300 transition">FAQ</a></li>
              <li><a href="#" class="hover:text-blue-300 transition">Common Issues</a></li>
              <li><a href="#" class="hover:text-blue-300 transition">Support</a></li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  `
})
export class DocsComponent {}

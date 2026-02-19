import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-features',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-purple-950">
      <div class="container mx-auto px-6 py-12">
        <h1 class="text-4xl font-bold text-white mb-8">Features</h1>
        <div class="grid md:grid-cols-3 gap-8 mt-8">
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8 hover:border-blue-500 transition">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Portfolio Tracking</h3>
            <p class="text-slate-300">Real-time portfolio monitoring and analytics with detailed breakdown of your investments.</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8 hover:border-blue-500 transition">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Secure Transactions</h3>
            <p class="text-slate-300">Enterprise-grade security for all your financial transactions and data.</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8 hover:border-blue-500 transition">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Advanced Analytics</h3>
            <p class="text-slate-300">Comprehensive insights and charts to make informed investment decisions.</p>
          </div>
        </div>
      </div>
    </div>
  `
})
export class FeaturesComponent {}

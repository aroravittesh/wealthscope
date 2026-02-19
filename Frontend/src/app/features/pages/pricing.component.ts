import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-pricing',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-purple-950">
      <div class="container mx-auto px-6 py-12">
        <h1 class="text-4xl font-bold text-white mb-8">Pricing Plans</h1>
        <div class="grid md:grid-cols-3 gap-8 mt-8">
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-white mb-4">Free</h3>
            <p class="text-slate-300 mb-4">Perfect for getting started</p>
            <p class="text-3xl font-bold text-blue-300 mb-6">$0<span class="text-lg">/mo</span></p>
            <ul class="text-slate-300 space-y-2 mb-6">
              <li>✓ Basic portfolio tracking</li>
              <li>✓ Limited analytics</li>
              <li>✗ Advanced features</li>
            </ul>
          </div>
          <div class="bg-slate-800/50 border border-purple-500 rounded-lg p-8 ring-2 ring-purple-500/30">
            <h3 class="text-2xl font-bold text-white mb-4">Pro</h3>
            <p class="text-slate-300 mb-4">Most popular</p>
            <p class="text-3xl font-bold text-purple-300 mb-6">$9.99<span class="text-lg">/mo</span></p>
            <ul class="text-slate-300 space-y-2 mb-6">
              <li>✓ Full portfolio tracking</li>
              <li>✓ Advanced analytics</li>
              <li>✓ Premium support</li>
            </ul>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-white mb-4">Enterprise</h3>
            <p class="text-slate-300 mb-4">For professionals</p>
            <p class="text-3xl font-bold text-blue-300 mb-6">Custom</p>
            <ul class="text-slate-300 space-y-2 mb-6">
              <li>✓ Everything in Pro</li>
              <li>✓ API access</li>
              <li>✓ Dedicated support</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  `
})
export class PricingComponent {}

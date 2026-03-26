import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-terms',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-purple-950">
      <div class="container mx-auto px-6 py-12 max-w-3xl">
        <h1 class="text-4xl font-bold text-white mb-8">Terms of Service</h1>
        <div class="space-y-8 text-slate-300">
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Acceptance of Terms</h3>
            <p>By using FinSight, you agree to comply with these terms and conditions.</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">User Responsibilities</h3>
            <p>You are responsible for maintaining the confidentiality of your account and password, and for all activities that occur under your account.</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Disclaimer</h3>
            <p>FinSight provides investment tracking and analytics tools. We do not provide financial advice, and all investment decisions are your own responsibility.</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Limitation of Liability</h3>
            <p>FinSight shall not be liable for any indirect, incidental, or consequential damages resulting from use of our services.</p>
          </div>
        </div>
      </div>
    </div>
  `
})
export class TermsComponent {}

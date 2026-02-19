import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-privacy',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-purple-950">
      <div class="container mx-auto px-6 py-12 max-w-3xl">
        <h1 class="text-4xl font-bold text-white mb-8">Privacy Policy</h1>
        <div class="space-y-8 text-slate-300">
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Information We Collect</h3>
            <p>We collect information necessary to provide and improve our services, including account information, portfolio data, and usage analytics.</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">How We Use Your Data</h3>
            <p>Your data is used solely to provide our services, improve user experience, and maintain security. We never sell or share personal data with third parties.</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Data Security</h3>
            <p>We employ industry-standard security measures including encryption, two-factor authentication, and regular security audits.</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Your Rights</h3>
            <p>You have the right to access, modify, or delete your personal data at any time. Contact us for more information.</p>
          </div>
        </div>
      </div>
    </div>
  `
})
export class PrivacyComponent {}

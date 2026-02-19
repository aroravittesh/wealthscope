import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-security',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-purple-950">
      <div class="container mx-auto px-6 py-12">
        <h1 class="text-4xl font-bold text-white mb-8">Security</h1>
        <div class="max-w-2xl space-y-8">
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">🔐 Encryption</h3>
            <p class="text-slate-300">All data is encrypted in transit using industry-standard TLS 1.3 protocol.</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">🛡️ Two-Factor Authentication</h3>
            <p class="text-slate-300">Protect your account with optional 2FA including authenticator apps and SMS.</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">🔒 PCI Compliance</h3>
            <p class="text-slate-300">We are fully PCI DSS compliant ensuring your payment information is secure.</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">📝 Privacy Policy</h3>
            <p class="text-slate-300">Your privacy is our priority. We never share or sell your personal data.</p>
          </div>
        </div>
      </div>
    </div>
  `
})
export class SecurityComponent {}

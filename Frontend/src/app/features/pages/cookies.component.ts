import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-cookies',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-purple-950">
      <div class="container mx-auto px-6 py-12 max-w-3xl">
        <h1 class="text-4xl font-bold text-white mb-8">Cookie Policy</h1>
        <div class="space-y-8 text-slate-300">
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">What Are Cookies</h3>
            <p>Cookies are small text files stored on your device that help us remember your preferences and improve your browsing experience.</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Types of Cookies</h3>
            <ul class="space-y-2">
              <li>✓ Essential cookies - Required for authentication</li>
              <li>✓ Performance cookies - Help us understand usage patterns</li>
              <li>✓ Preference cookies - Remember your settings</li>
            </ul>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Cookie Consent</h3>
            <p>You can control cookie preferences in your browser settings. Some cookies are essential for service functionality.</p>
          </div>
        </div>
      </div>
    </div>
  `
})
export class CookiesComponent {}

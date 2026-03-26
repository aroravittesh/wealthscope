import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-about',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-purple-950">
      <div class="container mx-auto px-6 py-12">
        <h1 class="text-4xl font-bold text-white mb-8">About Wealthscope</h1>
        <div class="max-w-3xl space-y-8">
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Our Mission</h3>
            <p class="text-slate-300">To democratize investment management and provide tools that empower investors to make informed financial decisions.</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Our Story</h3>
            <p class="text-slate-300">Founded in 2023,Wealthscope was created by a team of financial experts and software engineers who believed that sophisticated investment tools should be accessible to everyone.</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">Our Values</h3>
            <ul class="text-slate-300 space-y-2">
              <li>✓ Transparency in all operations</li>
              <li>✓ Security as a top priority</li>
              <li>✓ Continuous innovation</li>
              <li>✓ User-centric design</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  `
})
export class AboutComponent {}

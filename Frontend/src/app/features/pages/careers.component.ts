import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-careers',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-purple-950">
      <div class="container mx-auto px-6 py-12">
        <h1 class="text-4xl font-bold text-white mb-8">Careers</h1>
        <div class="max-w-3xl space-y-8">
          <p class="text-slate-300 text-lg">Join our growing team and help shape the future of investment management.</p>
          <div class="space-y-4">
            <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8 hover:border-blue-500 transition">
              <h3 class="text-xl font-bold text-blue-300 mb-2">Senior Full Stack Engineer</h3>
              <p class="text-slate-400 mb-2">Remote • Full-time</p>
              <p class="text-slate-300">We're looking for an experienced engineer to help build our next-generation platform.</p>
            </div>
            <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8 hover:border-blue-500 transition">
              <h3 class="text-xl font-bold text-blue-300 mb-2">Product Manager</h3>
              <p class="text-slate-400 mb-2">San Francisco • Full-time</p>
              <p class="text-slate-300">Lead product strategy and work with our engineering and design teams.</p>
            </div>
            <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8 hover:border-blue-500 transition">
              <h3 class="text-xl font-bold text-blue-300 mb-2">Financial Analyst</h3>
              <p class="text-slate-400 mb-2">New York • Full-time</p>
              <p class="text-slate-300">Contribute to our financial models and investment research division.</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  `
})
export class CareersComponent {}

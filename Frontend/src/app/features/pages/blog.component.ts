import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-blog',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-purple-950">
      <div class="container mx-auto px-6 py-12">
        <h1 class="text-4xl font-bold text-white mb-8">Blog</h1>
        <div class="grid md:grid-cols-2 gap-8">
          <article class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8 hover:border-blue-500 transition">
            <h3 class="text-2xl font-bold text-blue-300 mb-2">Getting Started with FinSight</h3>
            <p class="text-slate-400 text-sm mb-4">3 days ago</p>
            <p class="text-slate-300">Learn how to set up your portfolio and start tracking your investments with FinSight.</p>
          </article>
          <article class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8 hover:border-blue-500 transition">
            <h3 class="text-2xl font-bold text-blue-300 mb-2">Investment Tips for Beginners</h3>
            <p class="text-slate-400 text-sm mb-4">1 week ago</p>
            <p class="text-slate-300">Essential principles every beginner investor should know about diversification and risk management.</p>
          </article>
          <article class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8 hover:border-blue-500 transition">
            <h3 class="text-2xl font-bold text-blue-300 mb-2">Market Analysis Report</h3>
            <p class="text-slate-400 text-sm mb-4">2 weeks ago</p>
            <p class="text-slate-300">Monthly analysis of market trends and what they mean for your investment strategy.</p>
          </article>
          <article class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8 hover:border-blue-500 transition">
            <h3 class="text-2xl font-bold text-blue-300 mb-2">New Features Release</h3>
            <p class="text-slate-400 text-sm mb-4">3 weeks ago</p>
            <p class="text-slate-300">Introducing new advanced analytics and custom report generation features.</p>
          </article>
        </div>
      </div>
    </div>
  `
})
export class BlogComponent {}

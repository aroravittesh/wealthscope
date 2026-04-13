import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { PortfolioService } from '../services/portfolio.service';
import { Observable } from 'rxjs';
import { DashboardMetrics } from '../models';

@Component({
  selector: 'app-reporting-analytics',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="min-h-screen p-8 text-slate-200 font-sans relative overflow-hidden">
      <!-- Background Ambient Glows -->
      <div class="absolute top-[-10%] left-[-10%] w-96 h-96 bg-blue-600/20 rounded-full blur-[120px] pointer-events-none"></div>
      <div class="absolute bottom-[-10%] right-[-10%] w-96 h-96 bg-purple-600/20 rounded-full blur-[120px] pointer-events-none"></div>
      
      <div class="max-w-4xl mx-auto relative z-10 space-y-8">
        <h2 class="text-4xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-purple-500 tracking-tight">Reporting & Analytics</h2>
        
        <div class="flex flex-wrap gap-4">
          <button (click)="downloadPDF()" class="bg-gradient-to-r from-blue-600 to-indigo-600 text-white font-semibold px-6 py-2.5 rounded-xl shadow-[0_4px_14px_0_rgba(79,70,229,0.39)] hover:shadow-[0_6px_20px_rgba(79,70,229,0.23)] hover:scale-105 transition-all duration-300 transform">
            <span class="flex items-center gap-2">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>
              Download PDF
            </span>
          </button>
          <button (click)="exportCSV()" class="bg-slate-800 border border-slate-700 hover:bg-slate-700 text-white font-semibold px-6 py-2.5 rounded-xl transition-all duration-300 hover:scale-105 transform">
            Export CSV
          </button>
        </div>

        <ng-container *ngIf="metrics$ | async as metrics; else loadingState">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            
            <!-- Portfolio Analytics Card -->
            <div class="bg-slate-900/50 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 shadow-2xl relative overflow-hidden group hover:border-blue-500/50 transition-colors duration-500">
              <div class="absolute inset-0 bg-gradient-to-br from-blue-500/5 to-purple-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
              <h3 class="text-xl font-bold text-white mb-6 flex items-center gap-2">
                <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path></svg>
                Key Metrics
              </h3>
              
              <div class="space-y-4 relative z-10">
                <div class="flex justify-between items-center bg-slate-800/50 p-3 rounded-lg border border-slate-700/30">
                  <span class="text-slate-400 text-sm">Total Value</span>
                  <span class="font-bold text-lg text-white">\${{ metrics.totalPortfolioValue | number:'1.2-2' }}</span>
                </div>
                <div class="flex justify-between items-center bg-slate-800/50 p-3 rounded-lg border border-slate-700/30">
                  <span class="text-slate-400 text-sm">Profit / Loss</span>
                  <span class="font-bold text-lg" [ngClass]="{'text-green-400': metrics.totalProfitLoss > 0, 'text-red-400': metrics.totalProfitLoss < 0}">
                    {{ metrics.totalProfitLoss > 0 ? '+' : '' }}\${{ metrics.totalProfitLoss | number:'1.2-2' }}
                  </span>
                </div>
                <div class="flex justify-between items-center bg-slate-800/50 p-3 rounded-lg border border-slate-700/30">
                  <span class="text-slate-400 text-sm">P/L %</span>
                  <span class="font-bold text-lg" [ngClass]="{'text-green-400': metrics.profitLossPercentage > 0, 'text-red-400': metrics.profitLossPercentage < 0}">
                    {{ metrics.profitLossPercentage > 0 ? '+' : '' }}{{ metrics.profitLossPercentage | number:'1.2-2' }}%
                  </span>
                </div>
              </div>
            </div>

            <!-- Asset Allocation CSS Chart -->
            <div class="bg-slate-900/50 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 shadow-2xl relative overflow-hidden group hover:border-purple-500/50 transition-colors duration-500">
              <h3 class="text-xl font-bold text-white mb-6 flex items-center gap-2">
                <svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.488 9H15V3.512A9.025 9.025 0 0120.488 9z"></path></svg>
                Asset Allocation
              </h3>
              
              <div class="mb-4">
                <div class="w-full h-4 bg-slate-800 rounded-full overflow-hidden flex shadow-inner hover:scale-105 transition-transform cursor-pointer">
                  <div class="h-full bg-blue-500" style="width: 60%" title="Stocks: 60%"></div>
                  <div class="h-full bg-purple-500" style="width: 30%" title="Bonds: 30%"></div>
                  <div class="h-full bg-emerald-500" style="width: 10%" title="Cash: 10%"></div>
                </div>
              </div>
              
              <div class="grid grid-cols-3 gap-2 text-center text-sm mt-6">
                <div class="hover:-translate-y-1 transition-transform cursor-default">
                  <div class="flex items-center justify-center gap-1"><div class="w-2 h-2 rounded-full bg-blue-500"></div><span class="text-slate-300">Stocks</span></div>
                  <span class="font-bold text-white text-lg">60%</span>
                </div>
                <div class="hover:-translate-y-1 transition-transform cursor-default">
                  <div class="flex items-center justify-center gap-1"><div class="w-2 h-2 rounded-full bg-purple-500"></div><span class="text-slate-300">Bonds</span></div>
                  <span class="font-bold text-white text-lg">30%</span>
                </div>
                <div class="hover:-translate-y-1 transition-transform cursor-default">
                  <div class="flex items-center justify-center gap-1"><div class="w-2 h-2 rounded-full bg-emerald-500"></div><span class="text-slate-300">Cash</span></div>
                  <span class="font-bold text-white text-lg">10%</span>
                </div>
              </div>
            </div>

            <!-- AI Insights -->
            <div class="md:col-span-2 bg-slate-900/50 backdrop-blur-xl rounded-2xl p-6 border border-slate-700/50 shadow-2xl relative overflow-hidden group hover:border-emerald-500/50 transition-colors duration-500">
              <h3 class="text-xl font-bold text-white mb-4 flex items-center gap-2">
                <svg class="w-5 h-5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
                AI Portfolio Insights
              </h3>
              <div class="space-y-3">
                <div class="bg-emerald-500/10 border border-emerald-500/20 p-4 rounded-xl flex items-start gap-4 mb-2 hover:bg-emerald-500/20 transition-colors cursor-pointer">
                  <div class="bg-emerald-500/20 p-2 rounded-lg"><span class="text-xl">✨</span></div>
                  <div>
                    <h4 class="text-emerald-300 font-semibold text-sm mb-1">Diversification Good</h4>
                    <p class="text-slate-400 text-sm">Your portfolio is well diversified across sectors, matching your medium risk profile.</p>
                  </div>
                </div>
                <div class="bg-blue-500/10 border border-blue-500/20 p-4 rounded-xl flex items-start gap-4 hover:bg-blue-500/20 transition-colors cursor-pointer">
                  <div class="bg-blue-500/20 p-2 rounded-lg"><span class="text-xl">📈</span></div>
                  <div>
                    <h4 class="text-blue-300 font-semibold text-sm mb-1">Momentum Alert</h4>
                    <p class="text-slate-400 text-sm">Tech sector holdings are showing strong upward momentum. Consider holding rather than selling.</p>
                  </div>
                </div>
              </div>
            </div>

          </div>
        </ng-container>

        <!-- Loading State UI Glow-Up -->
        <ng-template #loadingState>
          <div class="h-[50vh] flex flex-col items-center justify-center space-y-8 animate-in fade-in duration-1000">
             <!-- Fancy spinning orb & logo -->
             <div class="relative w-28 h-28 hover:scale-105 transition-transform cursor-pointer group">
               <div class="absolute inset-0 bg-blue-500 rounded-full blur-[32px] opacity-40 group-hover:opacity-60 transition-opacity animate-pulse"></div>
               <div class="absolute inset-0 border-[3px] border-transparent border-t-blue-400 border-r-purple-500 rounded-full animate-spin shadow-[0_0_15px_rgba(59,130,246,0.5)]"></div>
               <div class="absolute inset-2 border-[3px] border-transparent border-b-cyan-400 border-l-blue-500 rounded-full animate-spin" style="animation-direction: reverse; animation-duration: 1.5s;"></div>
               <div class="absolute inset-0 flex items-center justify-center bg-slate-900/40 rounded-full backdrop-blur-sm border border-slate-700/50">
                  <svg class="w-10 h-10 transition-transform duration-500 group-hover:rotate-12" fill="none" viewBox="0 0 24 24">
                    <defs>
                      <linearGradient id="grad1" x1="0%" y1="0%" x2="100%" y2="100%">
                        <stop offset="0%" stop-color="#60a5fa" />
                        <stop offset="100%" stop-color="#a855f7" />
                      </linearGradient>
                    </defs>
                    <path stroke="url(#grad1)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" fill="url(#grad1)" fill-opacity="0.2"/>
                  </svg>
               </div>
             </div>
             
             <!-- Animated Text Processing -->
             <div class="text-center space-y-3">
               <h3 class="text-2xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-slate-200 to-slate-400 tracking-wide">
                 Synthesizing Analytics
               </h3>
               <div class="flex items-center justify-center gap-2 text-slate-400 text-sm font-medium">
                 <div class="w-2 h-2 rounded-full bg-blue-500 animate-ping"></div>
                 Establishing secure connection to portfolio servers...
               </div>
             </div>
          </div>
        </ng-template>

      </div>
    </div>
  `,
  styles: []
})
export class ReportingAnalyticsComponent implements OnInit {
  metrics$!: Observable<DashboardMetrics>;

  constructor(private portfolioService: PortfolioService) {}

  ngOnInit() {
    this.metrics$ = this.portfolioService.getPortfolioMetrics('1');
  }

  downloadPDF() {
    this.metrics$.subscribe(metrics => {
      // Very basic window.print() triggered PDF download implementation
      const pdfContent = `
        <html>
        <head>
          <title>Portfolio Report</title>
          <style>body { font-family: Arial, sans-serif; padding: 2em; }</style>
        </head>
        <body>
          <h1>Portfolio Analytics Report</h1>
          <p><strong>Total Value:</strong> $${metrics.totalPortfolioValue.toFixed(2)}</p>
          <p><strong>Profit/Loss:</strong> $${metrics.totalProfitLoss.toFixed(2)}</p>
          <p><strong>Day Change:</strong> ${metrics.profitLossPercentage}%</p>
          <p><strong>Total Invested:</strong> $${metrics.totalInvested.toFixed(2)}</p>
        </body>
        </html>
      `;
      const win = window.open('', '', 'width=800,height=600');
      if (win) {
        win.document.write(pdfContent);
        win.document.close();
        win.print();
        win.close();
      } else {
        alert('Please allow popups to generate the PDF report.');
      }
    });
  }
  
  exportCSV() {
    this.metrics$.subscribe(metrics => {
      const csvData = `Metric,Value\nTotal Value,$${metrics.totalPortfolioValue.toFixed(2)}\nProfit/Loss,${metrics.totalProfitLoss.toFixed(2)}\nDay Change,${metrics.profitLossPercentage}%\nTotal Invested,$${metrics.totalInvested.toFixed(2)}\n`;
      const blob = new Blob([csvData], { type: 'text/csv' });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'portfolio_summary.csv';
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    });
  }
}

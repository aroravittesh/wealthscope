import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-reporting-analytics',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="max-w-3xl mx-auto bg-slate-800 bg-opacity-60 rounded-xl p-8 mt-10 border border-slate-700 shadow-lg">
      <h2 class="text-2xl font-bold text-blue-300 mb-6">Reporting & Analytics (Stub)</h2>
      <div class="mb-6">
        <button class="bg-blue-600 text-white px-4 py-2 rounded mr-2">Download Portfolio Report (PDF)</button>
        <button class="bg-blue-600 text-white px-4 py-2 rounded mr-2">Email Portfolio Report</button>
        <button class="bg-blue-600 text-white px-4 py-2 rounded">Export Summary (CSV)</button>
      </div>
      <div class="bg-slate-900 rounded-lg p-6 border border-blue-700 mt-4">
        <h3 class="text-lg font-semibold text-blue-200 mb-2">Portfolio Analytics (Stub)</h3>
        <ul class="text-slate-400 text-sm list-disc ml-6">
          <li>Total Value: $123,456.78</li>
          <li>Profit/Loss: +$4,321.00</li>
          <li>Asset Allocation: 60% Stocks, 30% Bonds, 10% Cash</li>
          <li>Volatility Score: 0.23</li>
          <li>Diversification Score: 0.85</li>
          <li>Risk Classification: Medium</li>
        </ul>
      </div>
      <div class="bg-slate-900 rounded-lg p-6 border border-blue-700 mt-4">
        <h3 class="text-lg font-semibold text-blue-200 mb-2">AI Insights (Stub)</h3>
        <ul class="text-slate-400 text-sm list-disc ml-6">
          <li>"Your portfolio is well diversified for medium risk."</li>
          <li>"Consider rebalancing to reduce exposure to stocks."</li>
          <li>"AI Q&A: Ask about your portfolio!"</li>
        </ul>
      </div>
    </div>
  `,
  styles: []
})
export class ReportingAnalyticsComponent {}

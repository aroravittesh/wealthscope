import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-devops-testing',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="max-w-xl mx-auto bg-slate-800 bg-opacity-60 rounded-xl p-8 mt-10 border border-slate-700 shadow-lg">
      <h2 class="text-2xl font-bold text-purple-300 mb-6">DevOps & Testing (Stub)</h2>
      <div class="bg-slate-900 rounded-lg p-6 border border-purple-700 mt-4">
        <h3 class="text-lg font-semibold text-purple-200 mb-2">Testing Strategy</h3>
        <ul class="text-slate-400 text-sm list-disc ml-6">
          <li>Unit tests: <span class="text-purple-400">Planned</span></li>
          <li>Integration tests: <span class="text-purple-400">Planned</span></li>
          <li>Mock/test DB: <span class="text-purple-400">Planned</span></li>
          <li>API contract validation: <span class="text-purple-400">Planned</span></li>
        </ul>
      </div>
      <div class="bg-slate-900 rounded-lg p-6 border border-purple-700 mt-4">
        <h3 class="text-lg font-semibold text-purple-200 mb-2">DevOps & Deployment</h3>
        <ul class="text-slate-400 text-sm list-disc ml-6">
          <li>Dockerized backend: <span class="text-purple-400">Planned</span></li>
          <li>CI/CD pipeline: <span class="text-purple-400">Planned</span></li>
          <li>Environment-based configs: <span class="text-purple-400">Planned</span></li>
          <li>Cloud monitoring: <span class="text-purple-400">Planned</span></li>
          <li>Logging strategy: <span class="text-purple-400">Planned</span></li>
        </ul>
      </div>
    </div>
  `,
  styles: []
})
export class DevopsTestingComponent {}

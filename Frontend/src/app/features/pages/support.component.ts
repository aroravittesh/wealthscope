import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-support',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-purple-950">
      <div class="container mx-auto px-6 py-12">
        <h1 class="text-4xl font-bold text-white mb-8">Support</h1>
        <div class="grid md:grid-cols-2 gap-8">
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">📧 Email Support</h3>
            <p class="text-slate-300 mb-4">Get in touch with our support team</p>
            <p class="text-blue-300 font-mono">support&#64;finsight.com</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">💬 Live Chat</h3>
            <p class="text-slate-300 mb-4">Chat with our support team in real-time</p>
            <p class="text-slate-300">Available 24/7</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">📚 Knowledge Base</h3>
            <p class="text-slate-300 mb-4">Browse our comprehensive knowledge base</p>
            <p class="text-slate-300">Hundreds of articles and guides</p>
          </div>
          <div class="bg-slate-800/50 border border-blue-500/30 rounded-lg p-8">
            <h3 class="text-2xl font-bold text-blue-300 mb-4">🎓 Webinars</h3>
            <p class="text-slate-300 mb-4">Join our regular training sessions</p>
            <p class="text-slate-300">Weekly webinars on investing</p>
          </div>
        </div>
      </div>
    </div>
  `
})
export class SupportComponent {}

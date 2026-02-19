import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

@Component({
  selector: 'app-user-profile',
  standalone: true,
  imports: [CommonModule, FormsModule, ReactiveFormsModule],
  template: `
    <div class="max-w-lg mx-auto bg-slate-800 bg-opacity-60 rounded-xl p-8 mt-10 border border-slate-700 shadow-lg">
      <h2 class="text-2xl font-bold text-white mb-6">User Profile</h2>
      <form class="space-y-4">
        <div>
          <label class="block text-slate-300 mb-1">Full Name</label>
          <input class="w-full px-4 py-2 rounded bg-slate-700 border border-slate-600 text-white" placeholder="Full Name" />
        </div>
        <div>
          <label class="block text-slate-300 mb-1">Email</label>
          <input class="w-full px-4 py-2 rounded bg-slate-700 border border-slate-600 text-white" placeholder="Email" />
        </div>
        <div>
          <label class="block text-slate-300 mb-1">Phone</label>
          <input class="w-full px-4 py-2 rounded bg-slate-700 border border-slate-600 text-white" placeholder="Phone" />
        </div>
        <div>
          <label class="block text-slate-300 mb-1">Risk Preference</label>
          <select class="w-full px-4 py-2 rounded bg-slate-700 border border-slate-600 text-white">
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
          </select>
        </div>
        <button type="submit" class="w-full bg-blue-600 text-white py-2 rounded mt-4">Update Profile</button>
      </form>
    </div>
  `,
  styles: []
})
export class UserProfileComponent {}

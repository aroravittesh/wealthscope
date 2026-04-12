import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule, AbstractControl } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { AuthService } from '../../../services/auth.service';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterModule],
  template: `
    <div class="min-h-screen flex items-center justify-center p-4">
      <div class="w-full max-w-md">
        <div class="text-center mb-8">
          <div class="w-16 h-16 bg-gradient-to-r from-blue-500 to-purple-600 rounded-2xl mx-auto mb-4 flex items-center justify-center shadow-lg hover:shadow-blue-500/50 transition-shadow duration-300">
            <span class="text-2xl font-bold text-white">S</span>
          </div>
          <h1 class="text-3xl font-bold text-white mb-2">WealthScope</h1>
          <p class="text-slate-400">Create Your Account</p>
        </div>
        <div class="bg-slate-800 bg-opacity-50 backdrop-filter backdrop-blur-lg rounded-2xl p-8 border border-slate-700 shadow-2xl">
          <h2 class="text-2xl font-bold text-white mb-6">Sign Up</h2>
          <form [formGroup]="registerForm" (ngSubmit)="onSubmit()" class="space-y-4">
            <div>
              <label for="fullName" class="block text-sm font-medium text-slate-300 mb-2">Full Name</label>
              <input type="text" id="fullName" formControlName="fullName" placeholder="John Doe" class="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition" />
              <span *ngIf="registerForm.get('fullName')?.invalid && registerForm.get('fullName')?.touched" class="text-red-400 text-xs mt-1 block">Full name required</span>
            </div>
            <div>
              <label for="email" class="block text-sm font-medium text-slate-300 mb-2">Email Address</label>
              <input type="email" id="email" formControlName="email" placeholder="your@email.com" class="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition" />
              <span *ngIf="registerForm.get('email')?.invalid && registerForm.get('email')?.touched" class="text-red-400 text-xs mt-1 block">Valid email required</span>
            </div>
            <div>
              <label for="password" class="block text-sm font-medium text-slate-300 mb-2">Password</label>
              <input type="password" id="password" formControlName="password" placeholder="At least 8 characters" class="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition" />
              <span *ngIf="registerForm.get('password')?.invalid && registerForm.get('password')?.touched" class="text-red-400 text-xs mt-1 block">Password must be at least 8 characters</span>
            </div>
            <div>
              <label for="confirmPassword" class="block text-sm font-medium text-slate-300 mb-2">Confirm Password</label>
              <input type="password" id="confirmPassword" formControlName="confirmPassword" placeholder="••••••••" class="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition" />
              <span *ngIf="registerForm.get('confirmPassword')?.invalid && registerForm.get('confirmPassword')?.touched" class="text-red-400 text-xs mt-1 block">Passwords must match</span>
            </div>
            <label class="flex items-start">
              <input type="checkbox" formControlName="terms" class="w-4 h-4 rounded bg-slate-700 border-slate-600 accent-blue-500 cursor-pointer mt-1" />
              <span class="ml-3 text-sm text-slate-400">I agree to the <a href="#" class="text-blue-400 hover:text-blue-300 transition">Terms of Service</a> and <a href="#" class="text-blue-400 hover:text-blue-300 transition">Privacy Policy</a></span>
            </label>
            <button type="submit" [disabled]="registerForm.invalid || isLoading" class="w-full bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 text-white font-semibold py-3 rounded-lg transition disabled:opacity-50 disabled:cursor-not-allowed mt-6">
              <span *ngIf="!isLoading">Create Account</span>
              <span *ngIf="isLoading" class="flex items-center justify-center">
                <svg class="animate-spin -ml-1 mr-3 h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Creating account...
              </span>
            </button>
            <div *ngIf="errorMessage" class="bg-red-500 bg-opacity-20 border border-red-500 text-red-200 px-4 py-3 rounded-lg text-sm">{{ errorMessage }}</div>
          </form>
          <p class="text-center mt-6 text-slate-400">Already have an account? <a routerLink="/auth/login" class="text-blue-400 hover:text-blue-300 font-semibold transition">Sign in</a></p>
        </div>
      </div>
    </div>
  `,
  styles: []
})
export class RegisterComponent implements OnInit {
  registerForm!: FormGroup;
  isLoading = false;
  errorMessage = '';

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.initializeForm();
  }

  private initializeForm(): void {
    this.registerForm = this.fb.group({
      fullName: ['', [Validators.required]],
      email: ['', [Validators.required, Validators.email]],
      password: ['', [Validators.required, Validators.minLength(8)]],
      confirmPassword: ['', [Validators.required]],
      terms: [false, [Validators.requiredTrue]]
    }, {
      validators: this.passwordMatchValidator
    });
  }

  private passwordMatchValidator(control: AbstractControl): { [key: string]: any } | null {
    const password = control.get('password')?.value;
    const confirmPassword = control.get('confirmPassword')?.value;

    if (!password || !confirmPassword) {
      return null;
    }

    return password === confirmPassword ? null : { passwordMismatch: true };
  }

  onSubmit(): void {
    if (this.registerForm.invalid) return;

    this.isLoading = true;
    this.errorMessage = '';

    const { fullName, email, password } = this.registerForm.value;

    this.authService.register(email, password, fullName).subscribe({
      next: () => {
        this.isLoading = false;
        // Small delay to show success before navigation
        setTimeout(() => {
          this.router.navigate(['/dashboard']);
        }, 500);
      },
      error: (err) => {
        this.isLoading = false;
        this.errorMessage = err?.error?.message || 'Registration failed. Please try again.';
        console.error('Registration error:', err);
      },
      complete: () => {
        this.isLoading = false;
      }
    });
  }
}

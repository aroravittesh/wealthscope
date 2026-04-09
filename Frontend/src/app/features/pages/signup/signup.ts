import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { Router, RouterModule } from '@angular/router';
import { AuthService } from '../../../services/auth.service';
import {
  trigger,
  transition,
  style,
  animate
} from '@angular/animations';

@Component({
  selector: 'app-signup',
  standalone: true,
  imports: [
    FormsModule,
    CommonModule,
    RouterModule
  ],
  templateUrl: './signup.html',
  styleUrl: './signup.css',
  animations: [
    trigger('panelSlide', [
      transition(':enter', [
        style({
          transform: 'translateY(80px)',
          opacity: 0
        }),
        animate(
          '1500ms cubic-bezier(0.22, 1, 0.36, 1)',
          style({
            transform: 'translateY(0)',
            opacity: 1
          })
        )
      ]),
      transition(':leave', [
        animate(
          '1300ms cubic-bezier(0.4, 0, 0.2, 1)',
          style({
            transform: 'translateY(-60px)',
            opacity: 0
          })
        )
      ])
    ])
  ]
})
export class SignupComponent {

  email: string = '';
  password: string = '';
  riskPreference: string = '';
  isSubmitting = false;
  errorMessage: string | null = null;

  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  onSubmit(): void {
    if (!this.email || !this.password || !this.riskPreference) {
      return;
    }

    this.isSubmitting = true;
    this.errorMessage = null;

    this.authService.register(this.email, this.password, this.riskPreference)
      .subscribe({
        next: () => {
          this.isSubmitting = false;
          this.router.navigate(['/auth/login']);
        },
        error: (err) => {
          this.isSubmitting = false;
          this.errorMessage = err?.error?.message || err?.error || 'Registration failed. Please try again.';
        }
      });
  }

}
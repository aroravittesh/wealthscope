import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
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

  onSubmit(): void {
    console.log('Email:', this.email);
    console.log('Password:', this.password);
    console.log('Risk Preference:', this.riskPreference);
  }

}
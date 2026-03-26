// import { Component } from '@angular/core';
// import { FormsModule } from '@angular/forms';
// import { CommonModule } from '@angular/common';
// import { RouterModule } from '@angular/router';
// import {
//   trigger,
//   transition,
//   style,
//   animate
// } from '@angular/animations';

// @Component({
//   selector: 'app-login',
//   standalone: true,
//   imports: [
//     FormsModule,
//     CommonModule,
//     RouterModule   
//   ],
//   templateUrl: './login.html',
//   styleUrl: './login.css',
// animations: [
//   trigger('panelSlide', [
//     transition(':enter', [
//       style({
//         transform: 'translateY(80px)',
//         opacity: 0
//       }),
//       animate(
//         '1500ms cubic-bezier(0.22, 1, 0.36, 1)',
//         style({
//           transform: 'translateY(0)',
//           opacity: 1
//         })
//       )
//     ]),
//     transition(':leave', [
//       animate(
//         '1300ms cubic-bezier(0.4, 0, 0.2, 1)',
//         style({
//           transform: 'translateY(-60px)',
//           opacity: 0
//         })
//       )
//     ])
//   ])
// ]
// })
// export class LoginComponent {

//   email: string = '';
//   password: string = '';

//   onSubmit(): void {
//   }
// }
import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { RouterModule, Router } from '@angular/router';
import { AuthService } from '../../../services/auth.service';

import {
  trigger,
  transition,
  style,
  animate
} from '@angular/animations';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [
    FormsModule,
    CommonModule,
    RouterModule
  ],
  templateUrl: './login.html',
  styleUrls: ['./login.css'], // ✅ FIXED (was styleUrl)
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
export class LoginComponent {

  email: string = '';
  password: string = '';

  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  onSubmit(): void {

    if (!this.email || !this.password) {
      alert('Please enter email and password');
      return;
    }

    this.authService.login(this.email, this.password).subscribe({
      next: () => {
        console.log('Login successful');
        this.router.navigate(['/dashboard']);
      },
      error: (err) => {
        alert(err.error?.message || 'Login failed');
      }
    });
  }
}
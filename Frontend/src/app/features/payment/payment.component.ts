import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';

@Component({
  selector: 'app-payment',
  standalone: true,
  imports: [CommonModule, RouterModule, ReactiveFormsModule],
  templateUrl: './payment.component.html',
  styleUrl: './payment.component.scss'
})
export class PaymentComponent implements OnInit {
  selectedPlan: string = 'pro';
  selectedPaymentMethod: string = 'card';
  paymentForm!: FormGroup;

  paymentMethods = [
    { id: 'card', name: 'Credit/Debit Card', icon: '💳', description: 'Visa, Mastercard, American Express' },
    { id: 'ach', name: 'ACH Bank Transfer', icon: '🏦', description: 'US bank transfer (ACH)' },
    { id: 'paypal', name: 'PayPal', icon: '🌐', description: 'International payments' }
  ];

  planDetails: { [key: string]: any } = {
    free: { name: 'Free', price: 0, period: 'mo', duration: 'Forever free' },
    pro: { name: 'Pro', price: 9.99, period: 'mo', duration: 'Billing cycles' },
    enterprise: { name: 'Enterprise', price: 'Custom', period: '', duration: 'Custom terms' }
  };

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private fb: FormBuilder
  ) {
    this.initializeForm();
  }

  ngOnInit() {
    this.route.queryParams.subscribe(params => {
      if (params['plan']) {
        this.selectedPlan = params['plan'];
      }
    });
  }

  initializeForm() {
      this.paymentForm = this.fb.group({
        fullName: ['', [Validators.required, Validators.minLength(3)]],
        email: ['', [Validators.required, Validators.email]],
        cardNumber: ['', [Validators.required, Validators.pattern(/^\d{16}$/)]],
        cardExpiry: ['', [Validators.required, Validators.pattern(/^\d{2}\/\d{2}$/)]],
        cardCvv: ['', [Validators.required, Validators.pattern(/^\d{3}$/)]],
        achBankName: ['', Validators.required],
        achAccountNumber: ['', [Validators.required, Validators.minLength(10)]],
        paypalEmail: ['', [Validators.required, Validators.email]]
      });
  }

  selectPaymentMethod(method: string) {
    this.selectedPaymentMethod = method;
    this.paymentForm.reset();
  }

  processPayment() {
    if (this.selectedPlan === 'free') {
      this.completePayment();
      return;
    }

    // Validate appropriate fields based on payment method
    const requiredFields = this.getRequiredFields();
    let isValid = true;

    requiredFields.forEach(field => {
      const control = this.paymentForm.get(field);
      if (!control || !control.valid) {
        control?.markAsTouched();
        isValid = false;
      }
    });

    if (isValid) {
      this.completePayment();
    }
  }

  getRequiredFields(): string[] {
    const commonFields = ['fullName', 'email'];
    
    switch (this.selectedPaymentMethod) {
      case 'card':
        return [...commonFields, 'cardNumber', 'cardExpiry', 'cardCvv'];
      case 'ach':
        return [...commonFields, 'achBankName', 'achAccountNumber'];
      case 'paypal':
        return [...commonFields, 'paypalEmail'];
      default:
        return commonFields;
    }
  }

  completePayment() {
    // Simulate payment processing
    setTimeout(() => {
      alert(`Payment successful! You've subscribed to ${this.planDetails[this.selectedPlan].name} plan.`);
      this.router.navigate(['/dashboard']);
    }, 1500);
  }

  goBack() {
    this.router.navigate(['/']);
  }

  getFieldError(fieldName: string): string {
    const control = this.paymentForm.get(fieldName);
    if (control?.hasError('required')) {
      return `${fieldName.replace(/([A-Z])/g, ' $1')} is required`;
    }
    if (control?.hasError('email')) {
      return 'Invalid email format';
    }
    if (control?.hasError('minlength')) {
      return `Minimum length is ${control.getError('minlength')?.requiredLength}`;
    }
    if (control?.hasError('pattern')) {
      return 'Invalid format';
    }
    return '';
  }

  getPaymentMethodName(methodId: string): string {
    const method = this.paymentMethods.find(m => m.id === methodId);
    return method?.name || '';
  }
}

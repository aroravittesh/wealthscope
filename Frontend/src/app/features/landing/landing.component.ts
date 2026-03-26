import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, Router } from '@angular/router';

interface FeatureAnimation {
  name: string;
  title: string;
  description: string;
  color: string;
  icon: string;
}

@Component({
  selector: 'app-landing',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './landing.component.html',
  styleUrl: './landing.component.scss'
})
export class LandingComponent {
  activeFeature: FeatureAnimation | null = null;
  showAnimationModal = false;

  features: FeatureAnimation[] = [
    {
      name: 'analytics',
      title: 'Real-time Analytics',
      description: 'Monitor your portfolio performance with live updates and instant insights',
      color: 'from-blue-500 to-blue-600',
      icon: 'chart'
    },
    {
      name: 'ai-insights',
      title: 'AI Insights',
      description: 'Get intelligent recommendations based on your portfolio and market trends',
      color: 'from-purple-500 to-purple-600',
      icon: 'lightbulb'
    },
    {
      name: 'multi-portfolio',
      title: 'Multi-Portfolio',
      description: 'Create and manage multiple portfolios for different investment strategies',
      color: 'from-green-500 to-green-600',
      icon: 'briefcase'
    },
    {
      name: 'reports',
      title: 'Advanced Reports',
      description: 'Generate detailed reports and export your portfolio data anytime',
      color: 'from-yellow-500 to-yellow-600',
      icon: 'document'
    },
    {
      name: 'alerts',
      title: 'Smart Alerts',
      description: 'Receive notifications for price changes, portfolio thresholds, and opportunities',
      color: 'from-red-500 to-red-600',
      icon: 'bell'
    },
    {
      name: 'security',
      title: 'Enterprise Security',
      description: 'Bank-grade encryption and compliance with financial regulations',
      color: 'from-indigo-500 to-indigo-600',
      icon: 'lock'
    }
  ];

  constructor(private router: Router) {}

  goToPayment(plan: string) {
    this.router.navigate(['/payment'], { queryParams: { plan: plan.toLowerCase() } });
  }

  onFeatureClick(feature: FeatureAnimation) {
    this.activeFeature = feature;
    this.showAnimationModal = true;
  }

  closeAnimationModal() {
    this.showAnimationModal = false;
    setTimeout(() => {
      this.activeFeature = null;
    }, 300);
  }
}

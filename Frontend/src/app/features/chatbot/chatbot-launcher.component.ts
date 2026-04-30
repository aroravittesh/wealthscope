import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ChatbotPanelComponent } from './chatbot-panel.component';

@Component({
  selector: 'app-chatbot-launcher',
  standalone: true,
  imports: [CommonModule, ChatbotPanelComponent],
  templateUrl: './chatbot-launcher.component.html',
  styleUrl: './chatbot-launcher.component.scss'
})
export class ChatbotLauncherComponent {
  isOpen = false;

  toggleChat(): void {
    this.isOpen = !this.isOpen;
  }

  closeChat(): void {
    this.isOpen = false;
  }
}

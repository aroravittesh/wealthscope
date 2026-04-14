import { AfterViewChecked, Component, ElementRef, EventEmitter, Output, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { finalize } from 'rxjs/operators';
import { ChatbotService } from '../../services/chatbot.service';

interface ChatMessage {
  sender: 'user' | 'bot';
  text: string;
  createdAt: Date;
  error?: boolean;
}

@Component({
  selector: 'app-chatbot-panel',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './chatbot-panel.component.html',
  styleUrl: './chatbot-panel.component.scss'
})
export class ChatbotPanelComponent implements AfterViewChecked {
  @Output() close = new EventEmitter<void>();
  @ViewChild('messagesContainer') private messagesContainer?: ElementRef<HTMLDivElement>;

  readonly sessionId = 'demo-session-1';
  readonly title = 'Leo';
  readonly subtitle = 'WealthScope Assistant';

  messages: ChatMessage[] = [
    {
      sender: 'bot',
      text: 'Hi, I am Leo. Ask me about your portfolio, market trends, or WealthScope features.',
      createdAt: new Date()
    }
  ];

  draftMessage = '';
  isLoading = false;
  requestError = '';
  private shouldAutoScroll = false;

  constructor(private chatbotService: ChatbotService) {}

  ngAfterViewChecked(): void {
    if (this.shouldAutoScroll) {
      this.scrollToLatest();
      this.shouldAutoScroll = false;
    }
  }

  onClose(): void {
    this.close.emit();
  }

  onEnterSubmit(event: Event): void {
    const keyboardEvent = event as KeyboardEvent;
    if (keyboardEvent.shiftKey) {
      return;
    }
    keyboardEvent.preventDefault();
    this.sendMessage();
  }

  sendMessage(): void {
    const message = this.draftMessage.trim();
    if (!message || this.isLoading) {
      return;
    }

    this.requestError = '';
    this.pushMessage('user', message);
    this.draftMessage = '';
    this.isLoading = true;

    this.chatbotService
      .sendMessage({
        message,
        session_id: this.sessionId
      })
      .pipe(finalize(() => (this.isLoading = false)))
      .subscribe({
        next: (response) => {
          this.pushMessage('bot', response.response || 'I could not generate a response yet.');
        },
        error: () => {
          this.requestError = 'Unable to reach the assistant right now. Please try again.';
          this.pushMessage('bot', this.requestError, true);
        }
      });
  }

  trackByMessage(index: number): number {
    return index;
  }

  private pushMessage(sender: 'user' | 'bot', text: string, error = false): void {
    this.messages.push({
      sender,
      text,
      createdAt: new Date(),
      error
    });
    this.shouldAutoScroll = true;
  }

  private scrollToLatest(): void {
    const container = this.messagesContainer?.nativeElement;
    if (!container) {
      return;
    }
    container.scrollTop = container.scrollHeight;
  }
}

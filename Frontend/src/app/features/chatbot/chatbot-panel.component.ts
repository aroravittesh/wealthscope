import { AfterViewChecked, Component, ElementRef, EventEmitter, Output, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { finalize } from 'rxjs/operators';
import { ChatbotService } from '../../services/chatbot.service';
import { generateFollowUpSuggestions } from './follow-up-suggestions';

interface ChatMessage {
  sender: 'user' | 'bot';
  text: string;
  createdAt: Date;
  error?: boolean;
  followUps?: string[];
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
  readonly subtitle = 'Your portfolio & markets assistant';
  readonly suggestedPrompts: readonly string[] = [
    'Compare AAPL and MSFT',
    'Explain my portfolio risk',
    'Latest news on TSLA',
    'What is beta?',
    'Summarize my portfolio',
    'Explain diversification'
  ];

  messages: ChatMessage[] = [
    {
      sender: 'bot',
      text: 'Hi, I am Leo—your dashboard assistant. Ask about portfolios, market trends, or platform features.',
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
    this.clearMessageFollowUps();
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
          const botReply = response.response || 'I could not generate a response yet.';
          const followUps = generateFollowUpSuggestions({
            userMessage: message,
            botMessage: botReply
          });
          this.pushMessage('bot', botReply, false, followUps);
        },
        error: () => {
          this.requestError = 'Unable to reach the assistant right now. Please try again.';
          this.pushMessage('bot', this.requestError, true);
        }
      });
  }

  get showSuggestedPrompts(): boolean {
    return !this.isLoading && this.messages.every((message) => message.sender !== 'user');
  }

  onSelectSuggestedPrompt(prompt: string): void {
    if (this.isLoading) {
      return;
    }
    this.draftMessage = prompt;
    this.sendMessage();
  }

  onSelectFollowUp(prompt: string): void {
    if (this.isLoading) {
      return;
    }
    this.draftMessage = prompt;
    this.sendMessage();
  }

  trackByMessage(index: number): number {
    return index;
  }

  trackByParagraphIndex(index: number): number {
    return index;
  }

  trackBySuggestedPrompt(index: number): number {
    return index;
  }

  trackByFollowUp(index: number): number {
    return index;
  }

  isLatestBotMessage(index: number): boolean {
    for (let i = this.messages.length - 1; i >= 0; i--) {
      if (this.messages[i].sender === 'bot') {
        return i === index;
      }
    }
    return false;
  }

  /**
   * Splits assistant/user text into paragraphs for display. Double newlines
   * (blank lines) start a new paragraph; single newlines are preserved via
   * CSS `white-space: pre-line` on each paragraph.
   */
  splitMessageParagraphs(text: string): string[] {
    const raw = text ?? '';
    const normalized = raw.replace(/\r\n/g, '\n');
    const trimmed = normalized.trim();
    if (!trimmed) {
      return [''];
    }
    const blocks = normalized
      .split(/\n{2,}/)
      .map((s) => s.trim())
      .filter((s) => s.length > 0);
    return blocks.length > 0 ? blocks : [trimmed];
  }

  private pushMessage(sender: 'user' | 'bot', text: string, error = false, followUps?: string[]): void {
    this.messages.push({
      sender,
      text,
      createdAt: new Date(),
      error,
      followUps
    });
    this.shouldAutoScroll = true;
  }

  private clearMessageFollowUps(): void {
    this.messages = this.messages.map((message) => ({ ...message, followUps: undefined }));
  }

  private scrollToLatest(): void {
    const container = this.messagesContainer?.nativeElement;
    if (!container) {
      return;
    }
    container.scrollTop = container.scrollHeight;
  }
}

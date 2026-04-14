import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface ChatbotRequest {
  message: string;
  session_id: string;
}

export interface ChatbotResponse {
  response: string;
  session_id: string;
}

@Injectable({
  providedIn: 'root'
})
export class ChatbotService {
  private readonly chatBaseUrl = environment.chatApiUrl || environment.apiUrl;

  constructor(private http: HttpClient) {}

  sendMessage(payload: ChatbotRequest): Observable<ChatbotResponse> {
    return this.http.post<ChatbotResponse>(`${this.chatBaseUrl}/chat`, payload);
  }
}

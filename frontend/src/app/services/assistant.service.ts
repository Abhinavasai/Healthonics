import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';

export type AssistantChatResponse = {
  reply: string;
  provider: string;
  azure_configured: boolean;
  ollama_reachable: boolean;
  fallback_used: boolean;
};

@Injectable({ providedIn: 'root' })
export class AssistantService {
  constructor(private http: HttpClient) {}

  chat(message: string, context = ''): Observable<AssistantChatResponse> {
    return this.http.post<AssistantChatResponse>(ApiContract.assistant.chat, { message, context });
  }
}


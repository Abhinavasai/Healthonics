import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';

export interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
}

export interface AssistantChatResponse {
  reply: string;
  provider: string;
  azure_configured: boolean;
  ollama_reachable: boolean;
  fallback_used: boolean;
}

export interface SymptomCheckResponse {
  reply: string;
  provider: string;
  fallback_used: boolean;
  disclaimer: string;
}

export interface DrugInteractionResponse {
  reply: string;
  provider: string;
  count: number;
  fallback_used: boolean;
  disclaimer: string;
}

export interface SymptomTrendEntry {
  id: string;
  symptoms: string;
  urgency: 'emergency' | 'urgent' | 'routine' | 'selfcare';
  reply: string;
  provider: string;
  checked_at: string;
}

export interface SymptomTrendsResponse {
  checks: SymptomTrendEntry[];
  total: number;
  worsening: boolean;
}

export interface NoteAssistResponse {
  icd10_suggestions: string;
  soap_scaffold: string;
  clinical_flags: string;
  provider: string;
}

export interface PatientSummaryResponse {
  summary: string;
  provider: string;
  fallback_used: boolean;
  patient_id: string;
  patient_email: string;
}

@Injectable({ providedIn: 'root' })
export class AssistantService {
  constructor(private http: HttpClient) {}

  chat(message: string, history: ChatMessage[] = []): Observable<AssistantChatResponse> {
    return this.http.post<AssistantChatResponse>(ApiContract.assistant.chat, { message, history });
  }

  symptomCheck(symptoms: string, age?: number, gender?: string): Observable<SymptomCheckResponse> {
    return this.http.post<SymptomCheckResponse>(ApiContract.assistant.symptomCheck, { symptoms, age, gender });
  }

  drugInteractions(patientId: string): Observable<DrugInteractionResponse> {
    return this.http.post<DrugInteractionResponse>(ApiContract.assistant.drugInteractions, { patient_id: patientId });
  }

  patientSummary(patientId: string): Observable<PatientSummaryResponse> {
    return this.http.get<PatientSummaryResponse>(ApiContract.assistant.patientSummary(patientId));
  }

  noteAssist(appointmentId: string, noteText: string): Observable<NoteAssistResponse> {
    return this.http.post<NoteAssistResponse>(ApiContract.assistant.noteAssist(appointmentId), { note_text: noteText });
  }

  symptomTrends(): Observable<SymptomTrendsResponse> {
    return this.http.get<SymptomTrendsResponse>(ApiContract.assistant.symptomTrends);
  }
}

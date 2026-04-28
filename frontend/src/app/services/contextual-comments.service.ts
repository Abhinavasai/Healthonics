import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';

export type ContextType = 'record' | 'document' | 'knowledge_doc';
export type CommentVisibility = 'internal' | 'care_team' | 'patient_visible';

export interface ContextualCommentRow {
  id: string;
  context_type: ContextType;
  context_id: string;
  author_user_id: string;
  body: string;
  visibility: CommentVisibility;
  created_at: string;
}

@Injectable({ providedIn: 'root' })
export class ContextualCommentsService {
  constructor(private http: HttpClient) {}

  list(contextType: ContextType, contextId: string): Observable<{ comments: ContextualCommentRow[] }> {
    return this.http.get<{ comments: ContextualCommentRow[] }>(
      ApiContract.contextualComments.byContext(contextType, encodeURIComponent(contextId))
    );
  }

  create(
    contextType: ContextType,
    contextId: string,
    payload: { body: string; visibility: CommentVisibility }
  ): Observable<{ id: string; visibility: CommentVisibility }> {
    return this.http.post<{ id: string; visibility: CommentVisibility }>(
      ApiContract.contextualComments.byContext(contextType, encodeURIComponent(contextId)),
      payload
    );
  }
}


export type ApiListResponse<TItem, TKey extends string = 'items'> = {
  [P in TKey]: TItem[];
};

export interface AuditLogEntry {
  id: string;
  actor_user_id: string;
  action: string;
  entity_type: string;
  entity_id: string;
  detail: string;
  created_at: string;
}

export interface NotificationRow {
  id: string;
  title: string;
  body: string;
  channel: string;
  status: string;
  scheduled_for?: string | null;
  created_at: string;
}

export interface NotificationPreferenceRow {
  category: string;
  enabled: boolean;
  email_enabled: boolean;
  sms_enabled: boolean;
  in_app_enabled: boolean;
}

export type AdminLifecycleSettings = {
  new_user_window_days: number;
  inactive_window_days: number;
  updated_at: string;
  updated_by: string;
};

export type AdminLifecycleKpis = {
  total_users: number;
  total_patients: number;
  total_doctors: number;
  total_admins: number;
  new_users_in_window: number;
  inactive_users_count: number;
};

export type AdminLifecycleSettingsKpisResponse = {
  settings: AdminLifecycleSettings;
  kpis: AdminLifecycleKpis;
};

export type AdminAiRuntimeSettings = {
  ollama_enabled: boolean;
  fallback_enabled: boolean;
  rate_limit_enabled: boolean;
  rate_limit_per_minute: number;
  cache_enabled: boolean;
  cache_ttl_seconds: number;
  ollama_model: string;
  updated_at: string;
  updated_by: string;
};

export type AdminAiObservability = {
  queued_jobs_24h: number;
  completed_jobs_24h: number;
  failed_jobs_24h: number;
  avg_latency_seconds: number;
  cache_ready_count: number;
  pending_docs_count: number;
  failed_docs_count: number;
};

export type AdminAiRuntimeStatus = {
  ai_enabled: boolean;
  ollama_host: string;
  configured_model: string;
  ollama_reachable: boolean;
  model_available: boolean;
  active_mode: string;
};

export type AdminAiObservabilityResponse = {
  settings: AdminAiRuntimeSettings;
  observability: AdminAiObservability;
  runtime?: AdminAiRuntimeStatus;
};

export type AdminAiEvalResponse = {
  eval: {
    model_name: string;
    samples_evaluated: number;
    success_rate: number;
    quality_score: number;
  };
  runtime?: AdminAiRuntimeStatus;
};

export interface KnowledgeListDocument {
  id: string;
  title: string;
  excerpt: string;
  created_at: string;
  updated_at: string;
  current_version?: number;
  review_interval_days?: number;
  last_reviewed_at?: string;
  days_since_review?: number;
  health_score?: number;
  is_stale?: boolean;
}

export interface KnowledgeDocumentsListResponse {
  documents: KnowledgeListDocument[];
}

export interface KnowledgeDetail {
  id: string;
  title: string;
  body: string;
  updated_at: string;
  current_version: number;
  review_interval_days: number;
  last_reviewed_at?: string;
  days_since_review: number;
  health_score: number;
  is_stale: boolean;
}

export interface KnowledgeVersionRow {
  version: number;
  title: string;
  excerpt: string;
  created_at: string;
  created_by: string;
}

export interface KnowledgeVersionsResponse {
  versions: KnowledgeVersionRow[];
}

export interface SimilarityPair {
  similarity: number;
  document_a_id: string;
  document_b_id: string;
  title_a: string;
  title_b: string;
  chunk_a_index: number;
  chunk_b_index: number;
  excerpt_a: string;
  excerpt_b: string;
  conflict_hint: string;
}

export interface SimilarityScanResponse {
  threshold: number;
  chunks_scanned: number;
  pairs: SimilarityPair[];
}

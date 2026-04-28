import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import {
  AdminAiObservability,
  AdminAiRuntimeService,
  AdminAiRuntimeSettings
} from '../../services/admin-ai-runtime.service';

@Component({
  selector: 'app-admin-ai-observability',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './admin-ai-observability.component.html',
  styleUrl: './admin-ai-observability.component.scss'
})
export class AdminAiObservabilityComponent implements OnInit {
  loading = true;
  saving = false;
  runningEval = false;
  error: string | null = null;
  success: string | null = null;
  settings: AdminAiRuntimeSettings | null = null;
  obs: AdminAiObservability | null = null;
  evalResult: { model_name: string; samples_evaluated: number; success_rate: number; quality_score: number } | null =
    null;

  form = {
    ollama_enabled: false,
    fallback_enabled: true,
    rate_limit_enabled: true,
    rate_limit_per_minute: 30,
    cache_enabled: true,
    cache_ttl_seconds: 900,
    ollama_model: 'llama3.1:8b'
  };

  constructor(private readonly api: AdminAiRuntimeService) {}

  ngOnInit(): void {
    this.load();
  }

  load(): void {
    this.loading = true;
    this.error = null;
    this.api.getObservability().subscribe({
      next: (res) => {
        this.settings = res.settings;
        this.obs = res.observability;
        this.form = {
          ollama_enabled: res.settings.ollama_enabled,
          fallback_enabled: res.settings.fallback_enabled,
          rate_limit_enabled: res.settings.rate_limit_enabled,
          rate_limit_per_minute: res.settings.rate_limit_per_minute,
          cache_enabled: res.settings.cache_enabled,
          cache_ttl_seconds: res.settings.cache_ttl_seconds,
          ollama_model: res.settings.ollama_model
        };
        this.loading = false;
      },
      error: () => {
        this.error = 'Could not load AI observability.';
        this.loading = false;
      }
    });
  }

  save(): void {
    this.saving = true;
    this.error = null;
    this.success = null;
    this.api
      .updateSettings({
        ...this.form,
        rate_limit_per_minute: Number(this.form.rate_limit_per_minute),
        cache_ttl_seconds: Number(this.form.cache_ttl_seconds)
      })
      .subscribe({
        next: (res) => {
          this.settings = res.settings;
          this.obs = res.observability;
          this.success = 'AI runtime settings updated.';
          this.saving = false;
        },
        error: () => {
          this.error = 'Could not update AI runtime settings.';
          this.saving = false;
        }
      });
  }

  runEval(): void {
    this.runningEval = true;
    this.error = null;
    this.api.runEval().subscribe({
      next: (res) => {
        this.evalResult = res.eval;
        this.runningEval = false;
      },
      error: () => {
        this.error = 'Could not run AI eval.';
        this.runningEval = false;
      }
    });
  }
}

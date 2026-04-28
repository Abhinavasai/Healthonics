package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateSummaryWithAIRuntime_FallbackWhenAIDisabled(t *testing.T) {
	ConfigureAIRuntime(false, "http://127.0.0.1:11434", "llama3.1:8b", 500)
	settings := adminAIRuntimeSettings{
		OllamaEnabled:   true,
		FallbackEnabled: true,
		OllamaModel:     "llama3.1:8b",
	}
	summary, mode, err := GenerateSummaryWithAIRuntime(context.Background(), settings, "r.pdf", 4096, "sample extracted text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != "fallback_local" {
		t.Fatalf("expected fallback_local mode, got %s", mode)
	}
	if !strings.Contains(summary, "AI summary") {
		t.Fatalf("expected local summary fallback")
	}
}

func TestGenerateSummaryWithAIRuntime_OllamaSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/generate":
			_, _ = w.Write([]byte(`{"response":"Ollama summary output"}`))
		case "/api/tags":
			_, _ = w.Write([]byte(`{"models":[{"name":"llama3.1:8b"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ConfigureAIRuntime(true, server.URL, "llama3.1:8b", 1000)
	settings := adminAIRuntimeSettings{
		OllamaEnabled:   true,
		FallbackEnabled: true,
		OllamaModel:     "llama3.1:8b",
	}
	summary, mode, err := GenerateSummaryWithAIRuntime(context.Background(), settings, "r.pdf", 4096, "sample extracted text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != "ollama" {
		t.Fatalf("expected ollama mode, got %s", mode)
	}
	if summary != "Ollama summary output" {
		t.Fatalf("unexpected summary: %s", summary)
	}
}

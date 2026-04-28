package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type aiRuntimeConfig struct {
	AIEnabled       bool
	OllamaHost      string
	OllamaModel     string
	OllamaTimeoutMS int
}

var (
	aiRuntimeCfgMu sync.RWMutex
	aiRuntimeCfg   = aiRuntimeConfig{
		AIEnabled:       false,
		OllamaHost:      "http://127.0.0.1:11434",
		OllamaModel:     "llama3.1:8b",
		OllamaTimeoutMS: 7000,
	}
)

type ollamaRuntimeStatus struct {
	AIEnabled       bool   `json:"ai_enabled"`
	OllamaHost      string `json:"ollama_host"`
	ConfiguredModel string `json:"configured_model"`
	OllamaReachable bool   `json:"ollama_reachable"`
	ModelAvailable  bool   `json:"model_available"`
	ActiveMode      string `json:"active_mode"`
}

type ollamaTagsResponse struct {
	Models []struct {
		Name  string `json:"name"`
		Model string `json:"model"`
	} `json:"models"`
}

type ollamaGenerateResponse struct {
	Response string `json:"response"`
}

func ConfigureAIRuntime(enabled bool, host, model string, timeoutMS int) {
	aiRuntimeCfgMu.Lock()
	defer aiRuntimeCfgMu.Unlock()

	host = strings.TrimSpace(host)
	if host == "" {
		host = "http://127.0.0.1:11434"
	}
	model = strings.TrimSpace(model)
	if model == "" {
		model = "llama3.1:8b"
	}
	if timeoutMS <= 0 {
		timeoutMS = 7000
	}
	aiRuntimeCfg = aiRuntimeConfig{
		AIEnabled:       enabled,
		OllamaHost:      strings.TrimRight(host, "/"),
		OllamaModel:     model,
		OllamaTimeoutMS: timeoutMS,
	}
}

func getAIRuntimeConfig() aiRuntimeConfig {
	aiRuntimeCfgMu.RLock()
	defer aiRuntimeCfgMu.RUnlock()
	return aiRuntimeCfg
}

func getOllamaRuntimeStatus(ctx context.Context, settings adminAIRuntimeSettings) ollamaRuntimeStatus {
	cfg := getAIRuntimeConfig()
	out := ollamaRuntimeStatus{
		AIEnabled:       cfg.AIEnabled,
		OllamaHost:      cfg.OllamaHost,
		ConfiguredModel: modelFromSettingsOrConfig(settings, cfg),
		OllamaReachable: false,
		ModelAvailable:  false,
		ActiveMode:      "fallback",
	}
	if !cfg.AIEnabled || !settings.OllamaEnabled {
		return out
	}

	models, err := fetchOllamaModels(ctx, cfg)
	if err != nil {
		return out
	}
	out.OllamaReachable = true
	for _, m := range models {
		if m == out.ConfiguredModel {
			out.ModelAvailable = true
			out.ActiveMode = "ollama"
			return out
		}
	}
	return out
}

func GenerateSummaryWithAIRuntime(
	ctx context.Context,
	settings adminAIRuntimeSettings,
	filename string,
	sizeBytes int64,
	extracted string,
) (summary string, mode string, err error) {
	cfg := getAIRuntimeConfig()
	fallback := BuildSummaryFromExtractedTextForWorker(filename, sizeBytes, extracted)

	if !cfg.AIEnabled || !settings.OllamaEnabled {
		return fallback, "fallback_local", nil
	}
	model := modelFromSettingsOrConfig(settings, cfg)
	ollamaResp, genErr := callOllamaGenerate(ctx, cfg, model, extracted)
	if genErr != nil || strings.TrimSpace(ollamaResp) == "" {
		if settings.FallbackEnabled {
			return fallback, "fallback_local", nil
		}
		if genErr == nil {
			genErr = fmt.Errorf("ollama returned empty response")
		}
		return "", "ollama_error", genErr
	}
	return strings.TrimSpace(ollamaResp), "ollama", nil
}

func modelFromSettingsOrConfig(settings adminAIRuntimeSettings, cfg aiRuntimeConfig) string {
	model := strings.TrimSpace(settings.OllamaModel)
	if model == "" {
		model = strings.TrimSpace(cfg.OllamaModel)
	}
	if model == "" {
		model = "llama3.1:8b"
	}
	return model
}

func fetchOllamaModels(ctx context.Context, cfg aiRuntimeConfig) ([]string, error) {
	body, err := ollamaRequest(ctx, cfg, http.MethodGet, "/api/tags", nil)
	if err != nil {
		return nil, err
	}
	var tags ollamaTagsResponse
	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(tags.Models))
	for _, m := range tags.Models {
		if n := strings.TrimSpace(m.Name); n != "" {
			out = append(out, n)
			continue
		}
		if mm := strings.TrimSpace(m.Model); mm != "" {
			out = append(out, mm)
		}
	}
	return out, nil
}

func callOllamaGenerate(ctx context.Context, cfg aiRuntimeConfig, model, extracted string) (string, error) {
	prompt := fmt.Sprintf(
		"Summarize the following clinical document text in plain language with key findings and risks. Keep it concise and non-diagnostic.\n\n%s",
		extracted,
	)
	req := map[string]any{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	}
	body, err := ollamaRequest(ctx, cfg, http.MethodPost, "/api/generate", req)
	if err != nil {
		return "", err
	}
	var resp ollamaGenerateResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	return resp.Response, nil
}

func ollamaRequest(ctx context.Context, cfg aiRuntimeConfig, method, path string, payload any) ([]byte, error) {
	var bodyReader io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(raw)
	}
	timeout := time.Duration(cfg.OllamaTimeoutMS) * time.Millisecond
	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(rctx, method, cfg.OllamaHost+path, bodyReader)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("ollama %s %s failed: status=%d body=%s", method, path, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return raw, nil
}

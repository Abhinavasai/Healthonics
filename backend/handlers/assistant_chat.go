package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AssistantChatHandler struct {
	azureEndpoint   string
	azureAPIKey     string
	azureAPIVersion string
	azureModel      string
}

type assistantChatRequest struct {
	Message string `json:"message"`
	Context string `json:"context"`
}

func NewAssistantChatHandler(endpoint, apiKey, apiVersion, model string) *AssistantChatHandler {
	if strings.TrimSpace(apiVersion) == "" {
		apiVersion = "2025-01-01-preview"
	}
	if strings.TrimSpace(model) == "" {
		model = "gpt-4.1-mini"
	}
	return &AssistantChatHandler{
		azureEndpoint:   strings.TrimRight(strings.TrimSpace(endpoint), "/"),
		azureAPIKey:     strings.TrimSpace(apiKey),
		azureAPIVersion: strings.TrimSpace(apiVersion),
		azureModel:      strings.TrimSpace(model),
	}
}

func (h *AssistantChatHandler) Chat(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	var req assistantChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}
	prompt := "You are Healthonyx Care Assistant. Give concise, actionable, safe guidance for patients and clinicians. " +
		"If clinical uncertainty exists, suggest contacting a clinician."
	if ctx := strings.TrimSpace(req.Context); ctx != "" {
		prompt += "\n\nContext:\n" + ctx
	}

	if h.azureEndpoint != "" && h.azureAPIKey != "" {
		if out, err := h.callAzure(c.Request.Context(), prompt+"\n\nUser: "+msg); err == nil && strings.TrimSpace(out) != "" {
			c.JSON(http.StatusOK, gin.H{
				"reply":            strings.TrimSpace(out),
				"provider":         "azure_openai",
				"azure_configured": true,
				"ollama_reachable": h.ollamaReachable(c.Request.Context()),
				"fallback_used":    false,
			})
			return
		}
	}

	cfg := getAIRuntimeConfig()
	if cfg.AIEnabled {
		if out, err := callOllamaGenerate(c.Request.Context(), cfg, cfg.OllamaModel, prompt+"\n\nUser: "+msg); err == nil && strings.TrimSpace(out) != "" {
			c.JSON(http.StatusOK, gin.H{
				"reply":            strings.TrimSpace(out),
				"provider":         "ollama",
				"azure_configured": h.azureEndpoint != "" && h.azureAPIKey != "",
				"ollama_reachable": true,
				"fallback_used":    false,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"reply":            fallbackAssistantReply(msg),
		"provider":         "rule_fallback",
		"azure_configured": h.azureEndpoint != "" && h.azureAPIKey != "",
		"ollama_reachable": h.ollamaReachable(c.Request.Context()),
		"fallback_used":    true,
	})
}

func (h *AssistantChatHandler) callAzure(ctx context.Context, prompt string) (string, error) {
	payload := map[string]any{
		"model": h.azureModel,
		"input": prompt,
	}
	raw, _ := json.Marshal(payload)
	u := h.azureEndpoint + "/openai/responses?api-version=" + h.azureAPIVersion
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", h.azureAPIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return "", io.EOF
	}
	var parsed struct {
		OutputText string `json:"output_text"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	return parsed.OutputText, nil
}

func (h *AssistantChatHandler) ollamaReachable(ctx context.Context) bool {
	cfg := getAIRuntimeConfig()
	if !cfg.AIEnabled {
		return false
	}
	_, err := fetchOllamaModels(ctx, cfg)
	return err == nil
}

func fallbackAssistantReply(msg string) string {
	m := strings.ToLower(msg)
	switch {
	case strings.Contains(m, "appointment"):
		return "Open Appointments and choose your request to view status, comments, and follow-up actions."
	case strings.Contains(m, "message"):
		return "Messaging is available in the Messages page. If send fails, refresh once and retry."
	case strings.Contains(m, "document"), strings.Contains(m, "summary"):
		return "Use My documents for uploads and doctor view for AI summary actions."
	case strings.Contains(m, "hospital"), strings.Contains(m, "find care"):
		return "Use Find care with your location and department filter to locate nearby hospitals."
	default:
		return "I can help with messages, appointments, documents, notifications, and find care."
	}
}

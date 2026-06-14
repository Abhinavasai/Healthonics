package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// ChatMessage is one turn in the conversation (role: system/user/assistant).
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type assistantChatRequest struct {
	Message string        `json:"message"`
	History []ChatMessage `json:"history"`
}

func NewAssistantChatHandler(endpoint, apiKey, apiVersion, model string) *AssistantChatHandler {
	if strings.TrimSpace(apiVersion) == "" {
		apiVersion = "2025-03-01-preview"
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

const systemPrompt = `You are the Healthonyx Care Assistant — an intelligent, empathetic AI for patients and healthcare providers using the Healthonyx platform.

You can help with ANY healthcare-related topic, platform navigation, medical information, and general health questions. Be thorough, accurate, and warm.

Platform capabilities you know about:
- Appointments: book, view, cancel, reschedule; video teleconsult links; pre-visit questionnaires; check-in
- Prescriptions: view active prescriptions, request refills, download PDF
- Documents & Lab Results: upload documents, view lab results, AI interpretation
- Messages: secure messaging between patients and doctors; real-time chat
- Find Care: search nearby hospitals and doctors by location and specialty
- Health Monitoring: symptom checker, symptom trends, AI health summary
- Notifications: appointment reminders, medication alerts, lab result notifications
- Patient Dashboard: health overview, upcoming appointments, pending actions
- Doctor Portal: patient panel, waiting room queue, clinical notes, AI note assistant
- Admin Panel: audit logs, SLO dashboard, AI observability, user management

Guidelines:
1. Answer any healthcare or platform question fully and helpfully.
2. For medical questions, provide useful information while recommending professional consultation for diagnosis/treatment.
3. Give specific navigation hints when asked about app features (e.g. "Go to Appointments → click Book New").
4. Keep responses concise but complete — 2-5 sentences for simple questions, more detail for complex ones.
5. Never refuse to help with general health information. Only decline clearly harmful or illegal requests.
6. Be warm, professional, and supportive — many users are dealing with health concerns.
7. Remember conversation history and build on previous context.`

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

	messages := buildMessages(req.History, msg)
	ctx := c.Request.Context()

	azureConfigured := h.azureEndpoint != "" && h.azureAPIKey != ""

	// --- Primary: Azure OpenAI ---
	if azureConfigured {
		azureReply, azureErr := h.callAzureChatCompletions(ctx, messages)
		if azureErr == nil && strings.TrimSpace(azureReply) != "" {
			reply := strings.TrimSpace(azureReply)

			// --- Ollama validation: cross-check Azure answer quality ---
			cfg := getAIRuntimeConfig()
			validation := ""
			if cfg.AIEnabled {
				validation = h.ollamaValidate(ctx, cfg, msg, reply)
			}

			responsePayload := gin.H{
				"reply":             reply,
				"provider":         "azure_openai",
				"azure_configured": true,
				"fallback_used":    false,
			}
			if validation != "" {
				responsePayload["ollama_note"] = validation
			}
			c.JSON(http.StatusOK, responsePayload)
			return
		}
	}

	// --- Secondary: Ollama (local LLM) ---
	cfg := getAIRuntimeConfig()
	if cfg.AIEnabled {
		ollamaPrompt := systemPrompt + "\n\n" + historyToText(req.History) + "\nUser: " + msg + "\nAssistant:"
		ollamaReply, ollamaErr := callOllamaGenerate(ctx, cfg, cfg.OllamaModel, ollamaPrompt)
		if ollamaErr == nil && strings.TrimSpace(ollamaReply) != "" {
			c.JSON(http.StatusOK, gin.H{
				"reply":            strings.TrimSpace(ollamaReply),
				"provider":         "ollama",
				"azure_configured": azureConfigured,
				"fallback_used":    false,
			})
			return
		}
	}

	// --- Tertiary: Smart rule-based fallback ---
	c.JSON(http.StatusOK, gin.H{
		"reply":            smartFallbackReply(msg),
		"provider":         "rule_fallback",
		"azure_configured": azureConfigured,
		"fallback_used":    true,
	})
}

// ollamaValidate sends the Azure reply to Ollama for a brief quality check.
// Returns a short note if Ollama spots something to add; empty string otherwise.
func (h *AssistantChatHandler) ollamaValidate(ctx context.Context, cfg aiRuntimeConfig, userMsg, azureReply string) string {
	validateCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	prompt := fmt.Sprintf(`You are a medical AI quality reviewer. A user asked: "%s"
The AI responded: "%s"
In one sentence only: is this response accurate and safe? If it's good, say "Looks good." If there's an important correction or addition, state it briefly.`, userMsg, azureReply)

	note, err := callOllamaGenerate(validateCtx, cfg, cfg.OllamaModel, prompt)
	if err != nil {
		return ""
	}
	note = strings.TrimSpace(note)
	// Only surface the note if Ollama found something meaningful to add
	if strings.Contains(strings.ToLower(note), "looks good") || len(note) < 10 {
		return ""
	}
	return note
}

// buildMessages constructs the messages array for Chat Completions.
func buildMessages(history []ChatMessage, newUserMsg string) []ChatMessage {
	msgs := []ChatMessage{{Role: "system", Content: systemPrompt}}

	const maxHistory = 20
	start := 0
	if len(history) > maxHistory {
		start = len(history) - maxHistory
	}
	for _, h := range history[start:] {
		role := strings.ToLower(strings.TrimSpace(h.Role))
		if role != "user" && role != "assistant" {
			continue
		}
		msgs = append(msgs, ChatMessage{Role: role, Content: strings.TrimSpace(h.Content)})
	}
	msgs = append(msgs, ChatMessage{Role: "user", Content: newUserMsg})
	return msgs
}

func historyToText(history []ChatMessage) string {
	var sb strings.Builder
	for _, h := range history {
		role := strings.ToLower(strings.TrimSpace(h.Role))
		if role == "user" {
			sb.WriteString("User: ")
		} else if role == "assistant" {
			sb.WriteString("Assistant: ")
		} else {
			continue
		}
		sb.WriteString(strings.TrimSpace(h.Content))
		sb.WriteString("\n")
	}
	return sb.String()
}

// callAzureChatCompletions calls the Azure OpenAI Responses API (used by gpt-5.3-codex and newer models).
// Falls back to the Chat Completions API if the Responses API returns a non-2xx status.
func (h *AssistantChatHandler) callAzureChatCompletions(ctx context.Context, messages []ChatMessage) (string, error) {
	// Try Responses API first (required for gpt-5.x-codex models).
	if out, err := h.callAzureResponsesAPI(ctx, messages); err == nil {
		return out, nil
	}
	// Fallback to Chat Completions API (for gpt-4.x and older deployments).
	return h.callAzureChatCompletionsLegacy(ctx, messages)
}

// callAzureResponsesAPI calls the Azure OpenAI Responses API for gpt-5.x-codex models.
// Tries deployment-specific path first, then falls back to the global /openai/responses path.
func (h *AssistantChatHandler) callAzureResponsesAPI(ctx context.Context, messages []ChatMessage) (string, error) {
	// Separate system instructions from the conversation input.
	var instructions string
	var input []map[string]any
	for _, m := range messages {
		if m.Role == "system" {
			instructions = m.Content
			continue
		}
		input = append(input, map[string]any{"role": m.Role, "content": m.Content})
	}

	payload := map[string]any{
		"model":              h.azureModel,
		"input":             input,
		"max_output_tokens": 1200, // minimum is 16; 1200 gives full responses
	}
	if instructions != "" {
		payload["instructions"] = instructions
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Try deployment-specific URL first, then global fallback.
	urls := []string{
		h.azureEndpoint + "/openai/deployments/" + h.azureModel + "/responses?api-version=" + h.azureAPIVersion,
		h.azureEndpoint + "/openai/responses?api-version=" + h.azureAPIVersion,
	}

	hc := &http.Client{Timeout: 30 * time.Second}
	var lastErr error
	for _, u := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("api-key", h.azureAPIKey)

		resp, err := hc.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		resp.Body.Close()

		if resp.StatusCode >= http.StatusBadRequest {
			lastErr = fmt.Errorf("azure responses api %s status %d: %s", u, resp.StatusCode, string(body))
			continue
		}

		// Response: { "output_text": "...", "output": [{"content": [{"text": "..."}]}] }
		var parsed struct {
			OutputText string `json:"output_text"`
			Output     []struct {
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
			} `json:"output"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			lastErr = err
			continue
		}
		if strings.TrimSpace(parsed.OutputText) != "" {
			return parsed.OutputText, nil
		}
		for _, o := range parsed.Output {
			for _, c := range o.Content {
				if strings.TrimSpace(c.Text) != "" {
					return c.Text, nil
				}
			}
		}
		lastErr = fmt.Errorf("empty response from Azure Responses API")
	}
	return "", lastErr
}

// callAzureChatCompletionsLegacy calls the classic Chat Completions endpoint (gpt-4.x models).
func (h *AssistantChatHandler) callAzureChatCompletionsLegacy(ctx context.Context, messages []ChatMessage) (string, error) {
	payload := map[string]any{
		"messages":    messages,
		"max_tokens":  1200,
		"temperature": 0.7,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	u := h.azureEndpoint + "/openai/deployments/" + h.azureModel +
		"/chat/completions?api-version=" + h.azureAPIVersion

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", h.azureAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("azure chat completions status %d: %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	return parsed.Choices[0].Message.Content, nil
}

func (h *AssistantChatHandler) ollamaReachable(ctx context.Context) bool {
	cfg := getAIRuntimeConfig()
	if !cfg.AIEnabled {
		return false
	}
	_, err := fetchOllamaModels(ctx, cfg)
	return err == nil
}

// smartFallbackReply gives a helpful, detailed reply without any AI.
func smartFallbackReply(msg string) string {
	m := strings.ToLower(msg)

	switch {
	case contains(m, "appointment", "book", "schedule", "cancel", "reschedule"):
		return "To manage appointments: go to **Appointments** in the sidebar. To book a new one, use **Find Care** to search for a doctor, select an available slot, and confirm. Existing appointments can be cancelled or rescheduled from the appointment detail page."

	case contains(m, "prescription", "medication", "medicine", "refill", "drug"):
		return "Your prescriptions are in the **Prescriptions** section. You can view active medications, download PDFs, and request refills. Your doctor will be notified of refill requests and can approve or deny them. Always follow your doctor's instructions for medication changes."

	case contains(m, "document", "upload", "file", "lab", "result", "test"):
		return "Upload medical documents in **My Documents** or view lab results in the **Lab Results** section. Doctors can request AI-powered summaries of your documents. Lab results can be interpreted by our AI — tap the 'Interpret' button next to any lab result."

	case contains(m, "message", "chat", "contact", "send"):
		return "Use the **Messages** section to securely communicate with your care team. Select a thread to reply, or start a new conversation. Messages are end-to-end encrypted for your privacy."

	case contains(m, "hospital", "find care", "doctor", "specialist", "nearby", "location"):
		return "Use **Find Care** to search for hospitals and doctors near you. You can filter by specialty, distance, and availability. Once you find a provider, you can view their open slots and book directly."

	case contains(m, "symptom", "feel", "sick", "pain", "health check"):
		return "Use the **Symptom Checker** to describe how you're feeling. Our AI will ask follow-up questions and provide guidance on urgency and next steps. For emergencies, always call emergency services immediately."

	case contains(m, "notification", "reminder", "alert"):
		return "Check your **Notifications** for appointment reminders, medication alerts, and lab result updates. You can manage notification preferences in your profile settings."

	case contains(m, "video", "teleconsult", "virtual", "online visit"):
		return "For teleconsultations, check your appointment detail page — your doctor will have added a video link. Click it to join the virtual visit. Most virtual visits use Jitsi Meet which runs in your browser with no software to install."

	case contains(m, "password", "login", "account", "profile", "sign in"):
		return "To update your account details, use the profile section. If you've forgotten your password, use the 'Forgot password' option on the login page. Contact your administrator if you're locked out."

	case contains(m, "feature", "what can", "help", "how"):
		return "Healthonyx helps you manage your complete healthcare journey. Key features: **Appointments** (book & manage), **Prescriptions** (view & request refills), **Messages** (chat with your doctor), **Documents** (upload & view lab results), **Find Care** (locate doctors & hospitals), **Symptom Checker** (AI health assessment), and **Notifications** (reminders & alerts). What would you like help with specifically?"

	default:
		return "I'm your Healthonyx Care Assistant and I can help with appointments, prescriptions, lab results, messages, finding care, and general health questions. What do you need help with today?"
	}
}

func contains(s string, keywords ...string) bool {
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

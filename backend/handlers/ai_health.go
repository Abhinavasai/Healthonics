package handlers

// AI-powered health features — all use the same Azure OpenAI / Ollama / rule-based
// fallback chain already wired in assistant_chat.go.
//
// Endpoints:
//   POST /api/assistant/symptom-check        (patient) — triage guidance
//   POST /api/assistant/drug-interactions    (doctor)  — active prescription safety scan
//   GET  /api/patients/:patientId/ai-summary (doctor)  — clinical briefing from patient data

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
)

type AIHealthHandler struct {
	azureEndpoint   string
	azureAPIKey     string
	azureAPIVersion string
	azureModel      string
}

func NewAIHealthHandler(endpoint, apiKey, apiVersion, model string) *AIHealthHandler {
	if strings.TrimSpace(apiVersion) == "" {
		apiVersion = "2025-03-01-preview"
	}
	if strings.TrimSpace(model) == "" {
		model = "gpt-5.3-codex"
	}
	return &AIHealthHandler{
		azureEndpoint:   strings.TrimRight(strings.TrimSpace(endpoint), "/"),
		azureAPIKey:     strings.TrimSpace(apiKey),
		azureAPIVersion: strings.TrimSpace(apiVersion),
		azureModel:      strings.TrimSpace(model),
	}
}

// callAI sends a single-turn prompt through Azure → Ollama → rule fallback.
func (h *AIHealthHandler) callAI(c *gin.Context, systemMsg, userMsg string) (string, string) {
	msgs := []ChatMessage{
		{Role: "system", Content: systemMsg},
		{Role: "user", Content: userMsg},
	}

	azureHandler := &AssistantChatHandler{
		azureEndpoint:   h.azureEndpoint,
		azureAPIKey:     h.azureAPIKey,
		azureAPIVersion: h.azureAPIVersion,
		azureModel:      h.azureModel,
	}

	if h.azureEndpoint != "" && h.azureAPIKey != "" {
		if out, err := azureHandler.callAzureChatCompletions(c.Request.Context(), msgs); err == nil && strings.TrimSpace(out) != "" {
			return strings.TrimSpace(out), "azure_openai"
		}
	}

	cfg := getAIRuntimeConfig()
	if cfg.AIEnabled {
		prompt := systemMsg + "\n\n" + userMsg
		if out, err := callOllamaGenerate(c.Request.Context(), cfg, cfg.OllamaModel, prompt); err == nil && strings.TrimSpace(out) != "" {
			return strings.TrimSpace(out), "ollama"
		}
	}

	return "", "unavailable"
}

// ---------------------------------------------------------------------------
// Symptom Checker  POST /api/assistant/symptom-check
// ---------------------------------------------------------------------------

type symptomCheckRequest struct {
	Symptoms string `json:"symptoms"`
	Age      int    `json:"age"`
	Gender   string `json:"gender"`
}

func (h *AIHealthHandler) SymptomCheck(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Patients only"})
		return
	}

	var req symptomCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	symptoms := strings.TrimSpace(req.Symptoms)
	if symptoms == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symptoms is required"})
		return
	}
	if len([]rune(symptoms)) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symptoms must be 2000 characters or fewer"})
		return
	}

	systemMsg := `You are a medical triage assistant embedded in Healthonyx. Your role is to help patients understand the possible significance of their symptoms and guide them to the right level of care.

Rules:
- Never provide a definitive diagnosis.
- Always recommend professional evaluation for any concern.
- Classify urgency: Emergency (go to ER immediately), Urgent (see a doctor within 24 hours), Routine (schedule an appointment soon), Self-care (monitor at home, standard advice).
- Suggest 1-2 relevant medical specialties if the symptoms point clearly in a direction.
- Provide 2-3 self-care tips that are safe and evidence-based.
- Keep the response structured: Urgency, Possible Areas of Concern, Recommended Action, Self-Care Tips.
- Be empathetic and clear.`

	contextLine := ""
	if req.Age > 0 {
		contextLine += fmt.Sprintf("Patient age: %d. ", req.Age)
	}
	if strings.TrimSpace(req.Gender) != "" {
		contextLine += fmt.Sprintf("Gender: %s. ", strings.TrimSpace(req.Gender))
	}
	userMsg := contextLine + "Symptoms reported: " + symptoms

	reply, provider := h.callAI(c, systemMsg, userMsg)
	if reply == "" {
		reply = symptomFallback(symptoms)
		provider = "rule_fallback"
	}

	urgency := extractUrgency(reply)

	// Persist the check for trend tracking.
	if _, err := db.Pool.Exec(c.Request.Context(),
		`INSERT INTO symptom_checks (patient_id, symptoms, urgency, reply, provider, age, gender)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		claims.UserID, symptoms, urgency, reply, provider, req.Age, strings.TrimSpace(req.Gender),
	); err != nil {
		slog.Error("symptom check: failed to persist", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"reply":         reply,
		"provider":      provider,
		"urgency":       urgency,
		"fallback_used": provider == "rule_fallback",
		"disclaimer":    "This is not medical advice. Always consult a qualified healthcare professional.",
	})
}

func extractUrgency(reply string) string {
	lower := strings.ToLower(reply)
	switch {
	case strings.Contains(lower, "emergency"):
		return "emergency"
	case strings.Contains(lower, "urgent"):
		return "urgent"
	case strings.Contains(lower, "self-care") || strings.Contains(lower, "selfcare"):
		return "selfcare"
	default:
		return "routine"
	}
}

// SymptomTrends  GET /api/assistant/symptom-trends
func (h *AIHealthHandler) SymptomTrends(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Patients only"})
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id, symptoms, urgency, reply, provider, checked_at::text
		FROM symptom_checks
		WHERE patient_id = $1
		ORDER BY checked_at DESC
		LIMIT 30
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type entry struct {
		ID        string `json:"id"`
		Symptoms  string `json:"symptoms"`
		Urgency   string `json:"urgency"`
		Reply     string `json:"reply"`
		Provider  string `json:"provider"`
		CheckedAt string `json:"checked_at"`
	}
	var entries []entry
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.ID, &e.Symptoms, &e.Urgency, &e.Reply, &e.Provider, &e.CheckedAt); err == nil {
			entries = append(entries, e)
		}
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	// Compute worsening flag: emergency/urgent in latest 3 checks?
	worsening := false
	for i, e := range entries {
		if i >= 3 {
			break
		}
		if e.Urgency == "emergency" || e.Urgency == "urgent" {
			worsening = true
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"checks":    entries,
		"total":     len(entries),
		"worsening": worsening,
	})
}

func symptomFallback(symptoms string) string {
	s := strings.ToLower(symptoms)
	switch {
	case strings.Contains(s, "chest pain") || strings.Contains(s, "can't breathe") || strings.Contains(s, "cannot breathe"):
		return "Urgency: EMERGENCY. Chest pain or severe breathing difficulty requires immediate emergency evaluation. Please call emergency services or go to the nearest ER right away."
	case strings.Contains(s, "fever") && (strings.Contains(s, "rash") || strings.Contains(s, "stiff neck")):
		return "Urgency: EMERGENCY. Fever with rash or stiff neck may indicate a serious infection. Seek emergency care immediately."
	case strings.Contains(s, "fever"):
		return "Urgency: Urgent. A fever can indicate infection. Stay hydrated, rest, and monitor your temperature. If over 39.5°C (103°F) or lasting more than 3 days, see a doctor within 24 hours."
	case strings.Contains(s, "headache"):
		return "Urgency: Routine. Monitor your headache. If sudden and severe ('worst headache of your life'), go to the ER immediately. For typical headaches: rest, hydrate, and take OTC pain relief if appropriate."
	default:
		return "Urgency: Routine. Please describe your symptoms in detail to receive better guidance. For any concern, book an appointment through the Appointments section. If symptoms are severe or sudden, seek emergency care."
	}
}

// ---------------------------------------------------------------------------
// Drug Interaction Checker  POST /api/assistant/drug-interactions
// ---------------------------------------------------------------------------

type drugInteractionRequest struct {
	PatientID string `json:"patient_id"`
}

func (h *AIHealthHandler) DrugInteractions(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doctors and admins only"})
		return
	}

	var req drugInteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	patientID, err := uuid.Parse(strings.TrimSpace(req.PatientID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient_id"})
		return
	}

	// Fetch active prescriptions for the patient.
	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT medication_name, dosage, frequency, instructions
		FROM prescriptions
		WHERE patient_id = $1 AND status = 'active'
		ORDER BY created_at ASC
	`, patientID)
	if err != nil {
		slog.Error("drug interactions: failed to fetch prescriptions", "patient_id", patientID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type rxRow struct {
		Medication   string
		Dosage       string
		Frequency    string
		Instructions string
	}
	var prescriptions []rxRow
	for rows.Next() {
		var r rxRow
		if err := rows.Scan(&r.Medication, &r.Dosage, &r.Frequency, &r.Instructions); err != nil {
			slog.Error("drug interactions: scan error", "error", err)
			continue
		}
		prescriptions = append(prescriptions, r)
	}

	if len(prescriptions) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"reply":    "No active prescriptions found for this patient. Nothing to analyze.",
			"provider": "no_data",
			"count":    0,
		})
		return
	}

	// Build a structured list for the AI.
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("The patient has %d active prescription(s):\n\n", len(prescriptions)))
	for i, rx := range prescriptions {
		sb.WriteString(fmt.Sprintf("%d. %s — %s — %s", i+1, rx.Medication, rx.Dosage, rx.Frequency))
		if strings.TrimSpace(rx.Instructions) != "" {
			sb.WriteString(fmt.Sprintf(" (Notes: %s)", rx.Instructions))
		}
		sb.WriteString("\n")
	}

	systemMsg := `You are a clinical pharmacology assistant embedded in Healthonyx. A doctor is reviewing a patient's active prescriptions.

Your task:
1. Identify any clinically significant drug-drug interactions (rate each as: Major, Moderate, Minor, or None Found).
2. Flag any dosage concerns based on the frequency and drug class.
3. Highlight any duplicate therapeutic classes (e.g., two SSRIs, two NSAIDs).
4. Note any medications that commonly require monitoring (renal function, liver enzymes, electrolytes, INR, etc.).
5. Suggest any missing co-prescriptions that are standard of care (e.g., PPI with NSAIDs).

Format your response as clear sections. Be concise and clinically precise. Always note that this is an AI-assisted screening and final clinical judgment rests with the prescribing physician.`

	reply, provider := h.callAI(c, systemMsg, sb.String())
	if reply == "" {
		names := make([]string, 0, len(prescriptions))
		for _, rx := range prescriptions {
			names = append(names, rx.Medication)
		}
		reply = fmt.Sprintf("AI analysis unavailable. Patient has %d active prescription(s): %s. Please review interactions manually using a drug reference.", len(prescriptions), buildMedList(names))
		provider = "rule_fallback"
	}

	c.JSON(http.StatusOK, gin.H{
		"reply":        reply,
		"provider":     provider,
		"count":        len(prescriptions),
		"fallback_used": provider == "rule_fallback",
		"disclaimer":   "AI-assisted screening only. Clinical judgment of the prescribing physician takes precedence.",
	})
}

func buildMedList(meds []string) string {
	return strings.Join(meds, ", ")
}

// ---------------------------------------------------------------------------
// Note Assist  POST /api/appointments/:id/note-assist
// ---------------------------------------------------------------------------

type noteAssistRequest struct {
	NoteText string `json:"note_text"`
}

func (h *AIHealthHandler) NoteAssist(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doctors only"})
		return
	}

	apptID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment id"})
		return
	}

	var req noteAssistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	noteText := strings.TrimSpace(req.NoteText)
	if len([]rune(noteText)) > 4000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "note_text must be 4000 characters or fewer"})
		return
	}

	ctx := c.Request.Context()

	// Fetch appointment context.
	var reason, patientIDStr string
	var scheduledAt string
	if err := db.Pool.QueryRow(ctx, `
		SELECT reason, patient_id::text, scheduled_at::text
		FROM appointments WHERE id = $1
	`, apptID).Scan(&reason, &patientIDStr, &scheduledAt); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	patientID, _ := uuid.Parse(patientIDStr)

	// Fetch active prescriptions for context.
	rxRows, err := db.Pool.Query(ctx, `
		SELECT medication_name, dosage, frequency
		FROM prescriptions
		WHERE patient_id = $1 AND status = 'active'
		ORDER BY created_at DESC LIMIT 10
	`, patientID)
	var rxLines []string
	if err == nil {
		defer rxRows.Close()
		for rxRows.Next() {
			var med, dosage, freq string
			if err := rxRows.Scan(&med, &dosage, &freq); err == nil {
				rxLines = append(rxLines, fmt.Sprintf("%s %s %s", med, dosage, freq))
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Appointment reason: %s\n", reason))
	sb.WriteString(fmt.Sprintf("Scheduled: %s\n", scheduledAt[:10]))
	if len(rxLines) > 0 {
		sb.WriteString("Patient's active medications: " + strings.Join(rxLines, "; ") + "\n")
	} else {
		sb.WriteString("Patient's active medications: None on record.\n")
	}
	if noteText != "" {
		sb.WriteString("\nDraft clinical note from doctor:\n" + noteText)
	} else {
		sb.WriteString("\nNo draft note yet — generate a SOAP scaffold from the appointment reason.")
	}

	systemMsg := `You are a clinical documentation assistant embedded in Healthonyx, assisting a doctor in real-time as they write appointment notes.

Your output must always contain exactly three sections, each on its own line starting with the label:

ICD-10 Suggestions: (list 2-4 relevant ICD-10 codes with short descriptions, comma-separated, e.g. "J06.9 Acute upper respiratory infection, R05 Cough")
SOAP Scaffold: (a structured SOAP note outline with Subjective / Objective / Assessment / Plan filled in with reasonable starter content based on the reason and draft note)
Clinical Flags: (any drug interactions with current medications, monitoring requirements, or prescribing considerations relevant to this visit — or "None identified" if nothing relevant)

Be concise. Treat this as a real-time typing assistant — accuracy matters more than verbosity.`

	reply, provider := h.callAI(c, systemMsg, sb.String())

	icd10 := ""
	soap := ""
	flags := ""
	if reply != "" {
		for _, line := range strings.Split(reply, "\n") {
			switch {
			case strings.HasPrefix(line, "ICD-10 Suggestions:"):
				icd10 = strings.TrimSpace(strings.TrimPrefix(line, "ICD-10 Suggestions:"))
			case strings.HasPrefix(line, "SOAP Scaffold:"):
				soap = strings.TrimSpace(strings.TrimPrefix(line, "SOAP Scaffold:"))
			case strings.HasPrefix(line, "Clinical Flags:"):
				flags = strings.TrimSpace(strings.TrimPrefix(line, "Clinical Flags:"))
			}
		}
		// Fallback: if parsing failed (AI put content on multiple lines), return raw reply.
		if icd10 == "" && soap == "" {
			soap = reply
		}
	} else {
		icd10 = "AI unavailable — please code manually"
		soap = fmt.Sprintf("S: Patient presents with: %s\nO: (Examination findings)\nA: (Assessment)\nP: (Plan)", reason)
		flags = "AI unavailable — review medications manually"
		provider = "rule_fallback"
	}

	c.JSON(http.StatusOK, gin.H{
		"icd10_suggestions": icd10,
		"soap_scaffold":     soap,
		"clinical_flags":    flags,
		"provider":          provider,
	})
}

// ---------------------------------------------------------------------------
// AI Patient Summary  GET /api/patients/:patientId/ai-summary
// ---------------------------------------------------------------------------

func (h *AIHealthHandler) PatientSummary(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "doctor" && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doctors and admins only"})
		return
	}

	patientID, err := uuid.Parse(strings.TrimSpace(c.Param("patientId")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient id"})
		return
	}

	ctx := c.Request.Context()

	// Fetch patient email.
	var patientEmail string
	if err := db.Pool.QueryRow(ctx, `SELECT email FROM users WHERE id = $1`, patientID).Scan(&patientEmail); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
		return
	}

	// Fetch recent appointments (last 10).
	apptRows, err := db.Pool.Query(ctx, `
		SELECT scheduled_at::text, reason, status
		FROM appointments
		WHERE patient_id = $1
		ORDER BY scheduled_at DESC
		LIMIT 10
	`, patientID)
	if err != nil {
		slog.Error("patient summary: appointments query", "error", err)
	}
	var appts []string
	if apptRows != nil {
		defer apptRows.Close()
		for apptRows.Next() {
			var scheduledAt, reason, status string
			if err := apptRows.Scan(&scheduledAt, &reason, &status); err == nil {
				appts = append(appts, fmt.Sprintf("  - %s: %s (%s)", scheduledAt[:10], reason, status))
			}
		}
	}

	// Fetch active prescriptions.
	rxRows, err := db.Pool.Query(ctx, `
		SELECT medication_name, dosage, frequency, duration_days
		FROM prescriptions
		WHERE patient_id = $1 AND status = 'active'
		ORDER BY created_at DESC
	`, patientID)
	if err != nil {
		slog.Error("patient summary: prescriptions query", "error", err)
	}
	var rxList []string
	if rxRows != nil {
		defer rxRows.Close()
		for rxRows.Next() {
			var med, dosage, freq string
			var days int
			if err := rxRows.Scan(&med, &dosage, &freq, &days); err == nil {
				rxList = append(rxList, fmt.Sprintf("  - %s %s %s (%d days)", med, dosage, freq, days))
			}
		}
	}

	// Fetch recent document summaries (last 5 summarized).
	docRows, err := db.Pool.Query(ctx, `
		SELECT filename, summary
		FROM patient_documents
		WHERE patient_id = $1 AND summary IS NOT NULL AND summary_status = 'completed'
		ORDER BY created_at DESC
		LIMIT 5
	`, patientID)
	if err != nil {
		slog.Error("patient summary: documents query", "error", err)
	}
	var docSummaries []string
	if docRows != nil {
		defer docRows.Close()
		for docRows.Next() {
			var filename, summary string
			if err := docRows.Scan(&filename, &summary); err == nil {
				// Truncate long summaries for the prompt.
				runes := []rune(summary)
				if len(runes) > 300 {
					summary = string(runes[:300]) + "..."
				}
				docSummaries = append(docSummaries, fmt.Sprintf("  - %s: %s", filename, summary))
			}
		}
	}

	// Build the context block.
	var context strings.Builder
	context.WriteString(fmt.Sprintf("Patient: %s\n\n", patientEmail))

	if len(appts) > 0 {
		context.WriteString("Recent Appointments:\n")
		for _, a := range appts {
			context.WriteString(a + "\n")
		}
	} else {
		context.WriteString("Recent Appointments: None on record.\n")
	}

	context.WriteString("\n")
	if len(rxList) > 0 {
		context.WriteString("Active Prescriptions:\n")
		for _, r := range rxList {
			context.WriteString(r + "\n")
		}
	} else {
		context.WriteString("Active Prescriptions: None.\n")
	}

	context.WriteString("\n")
	if len(docSummaries) > 0 {
		context.WriteString("Document Summaries:\n")
		for _, d := range docSummaries {
			context.WriteString(d + "\n")
		}
	} else {
		context.WriteString("Document Summaries: No summarized documents.\n")
	}

	systemMsg := `You are a clinical briefing assistant for Healthonyx. A doctor is about to see a patient and needs a concise pre-visit briefing.

Generate a structured clinical briefing that includes:
1. **Patient Overview** — one sentence summary
2. **Active Issues** — what the patient is currently being treated for, based on prescriptions and appointment reasons
3. **Recent Visit History** — key appointments and their outcomes
4. **Medications at a Glance** — active prescriptions with relevant notes
5. **Flags & Recommendations** — anything that warrants attention (long gaps in care, medication duration concerns, follow-up needed)
6. **Suggested Discussion Points** — 2-3 things to consider discussing during the visit

Be concise, precise, and clinically useful. Write as if briefing a busy physician before a 15-minute consult.`

	reply, provider := h.callAI(c, systemMsg, context.String())
	if reply == "" {
		// Structured fallback without AI.
		var fb strings.Builder
		fb.WriteString(fmt.Sprintf("**Patient:** %s\n\n", patientEmail))
		if len(rxList) > 0 {
			fb.WriteString("**Active Prescriptions:**\n")
			for _, r := range rxList {
				fb.WriteString(r + "\n")
			}
			fb.WriteString("\n")
		}
		if len(appts) > 0 {
			fb.WriteString("**Recent Appointments:**\n")
			for _, a := range appts {
				fb.WriteString(a + "\n")
			}
		}
		fb.WriteString("\n_(AI summary unavailable — displaying raw data)_")
		reply = fb.String()
		provider = "rule_fallback"
	}

	c.JSON(http.StatusOK, gin.H{
		"summary":       reply,
		"provider":      provider,
		"fallback_used": provider == "rule_fallback",
		"patient_id":    patientID.String(),
		"patient_email": patientEmail,
	})
}

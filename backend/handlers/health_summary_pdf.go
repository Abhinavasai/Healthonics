package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/db"
)

type HealthSummaryPDFHandler struct{}

func NewHealthSummaryPDFHandler() *HealthSummaryPDFHandler {
	return &HealthSummaryPDFHandler{}
}

// healthSummaryTmpl is a minimal plain-text health summary (no paid PDF lib needed).
// The frontend receives text/plain and the browser downloads it as a .txt file.
// For a true PDF, a free library like pdfcpu or go-wkhtmltopdf can be swapped in later.
var healthSummaryTmpl = template.Must(template.New("summary").Parse(`
HEALTHONYX — PERSONAL HEALTH SUMMARY
======================================
Generated: {{ .GeneratedAt }}

PATIENT
-------
Email    : {{ .Email }}
Patient ID: {{ .PatientID }}

RECENT APPOINTMENTS (last 10)
------------------------------
{{ range .Appointments }}
  {{ .ScheduledAt }} | {{ .Status }} | {{ .Reason }}
{{ else }}
  No recent appointments.
{{ end }}

ACTIVE PRESCRIPTIONS
--------------------
{{ range .Prescriptions }}
  {{ .MedicationName }} | {{ .Dosage }} | {{ .Frequency }} | {{ .Instructions }}
{{ else }}
  No active prescriptions.
{{ end }}

RECENT SYMPTOM CHECKS (last 5)
--------------------------------
{{ range .SymptomChecks }}
  {{ .CheckedAt }} — {{ .Symptoms }} (Urgency: {{ .Urgency }})
{{ else }}
  No recent symptom checks.
{{ end }}

---
This document is for personal reference only and does not constitute medical advice.
`))

type healthSummaryData struct {
	GeneratedAt   string
	Email         string
	PatientID     string
	Appointments  []healthAppt
	Prescriptions []healthRx
	SymptomChecks []healthSymptom
}

type healthAppt struct {
	ScheduledAt string
	Status      string
	Reason      string
}

type healthRx struct {
	MedicationName string
	Dosage         string
	Frequency      string
	Instructions   string
}

type healthSymptom struct {
	CheckedAt string
	Symptoms  string
	Urgency   string
}

func (h *HealthSummaryPDFHandler) Export(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	ctx := c.Request.Context()

	// Patient email.
	var email string
	if err := db.Pool.QueryRow(ctx, `SELECT email FROM users WHERE id = $1`, claims.UserID).Scan(&email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	// Recent appointments.
	apptRows, err := db.Pool.Query(ctx, `
		SELECT scheduled_at::text, status, reason FROM appointments
		WHERE patient_id = $1 ORDER BY scheduled_at DESC LIMIT 10
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer apptRows.Close()
	var appts []healthAppt
	for apptRows.Next() {
		var a healthAppt
		if err := apptRows.Scan(&a.ScheduledAt, &a.Status, &a.Reason); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		appts = append(appts, a)
	}
	if err := apptRows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	// Active prescriptions.
	rxRows, err := db.Pool.Query(ctx, `
		SELECT medication_name, dosage, frequency, instructions FROM prescriptions
		WHERE patient_id = $1 AND status = 'active' ORDER BY created_at DESC
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rxRows.Close()
	var rxList []healthRx
	for rxRows.Next() {
		var r healthRx
		if err := rxRows.Scan(&r.MedicationName, &r.Dosage, &r.Frequency, &r.Instructions); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		rxList = append(rxList, r)
	}
	if err := rxRows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	// Recent symptom checks.
	symRows, err := db.Pool.Query(ctx, `
		SELECT checked_at::text, symptoms, urgency FROM symptom_checks
		WHERE patient_id = $1 ORDER BY checked_at DESC LIMIT 5
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer symRows.Close()
	var symList []healthSymptom
	for symRows.Next() {
		var s healthSymptom
		if err := symRows.Scan(&s.CheckedAt, &s.Symptoms, &s.Urgency); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		symList = append(symList, s)
	}
	if err := symRows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	data := healthSummaryData{
		GeneratedAt:   time.Now().UTC().Format("January 2, 2006 15:04 UTC"),
		Email:         email,
		PatientID:     claims.UserID.String(),
		Appointments:  appts,
		Prescriptions: rxList,
		SymptomChecks: symList,
	}

	var buf bytes.Buffer
	if err := healthSummaryTmpl.Execute(&buf, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	filename := fmt.Sprintf("healthonyx-summary-%s.txt", time.Now().UTC().Format("2006-01-02"))
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "text/plain; charset=utf-8", buf.Bytes())
}

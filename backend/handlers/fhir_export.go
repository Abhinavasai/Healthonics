package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/db"
)

// FHIRExportHandler builds a FHIR R4 Bundle for the requesting patient.
// No external dependencies — pure stdlib JSON construction.
type FHIRExportHandler struct{}

func NewFHIRExportHandler() *FHIRExportHandler { return &FHIRExportHandler{} }

// Export  GET /api/patient/fhir-export
func (h *FHIRExportHandler) Export(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Role != "patient" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Patients only"})
		return
	}

	ctx := c.Request.Context()
	patientID := claims.UserID.String()

	// Fetch patient details.
	var email string
	var createdAt string
	if err := db.Pool.QueryRow(ctx,
		`SELECT email, created_at::text FROM users WHERE id = $1`, claims.UserID,
	).Scan(&email, &createdAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	entries := []map[string]interface{}{}

	// --- Patient Resource ---
	entries = append(entries, map[string]interface{}{
		"fullUrl": "urn:uuid:" + patientID,
		"resource": map[string]interface{}{
			"resourceType": "Patient",
			"id":           patientID,
			"identifier": []map[string]interface{}{
				{"system": "https://healthonyx.app/patients", "value": patientID},
			},
			"telecom": []map[string]interface{}{
				{"system": "email", "value": email, "use": "home"},
			},
			"meta": map[string]interface{}{
				"lastUpdated": createdAt,
			},
		},
		"request": map[string]string{"method": "PUT", "url": "Patient/" + patientID},
	})

	// --- Appointments → FHIR Encounter resources ---
	apptRows, err := db.Pool.Query(ctx, `
		SELECT id::text, reason, status, scheduled_at::text, doctor_id::text
		FROM appointments WHERE patient_id = $1 ORDER BY scheduled_at DESC LIMIT 50
	`, claims.UserID)
	if err == nil {
		defer apptRows.Close()
		for apptRows.Next() {
			var id, reason, status, scheduledAt, doctorID string
			if err := apptRows.Scan(&id, &reason, &status, &scheduledAt, &doctorID); err != nil {
				continue
			}
			fhirStatus := appointmentStatusToFHIR(status)
			entries = append(entries, map[string]interface{}{
				"fullUrl": "urn:uuid:" + id,
				"resource": map[string]interface{}{
					"resourceType": "Encounter",
					"id":           id,
					"status":       fhirStatus,
					"class": map[string]string{
						"system":  "http://terminology.hl7.org/CodeSystem/v3-ActCode",
						"code":    "AMB",
						"display": "ambulatory",
					},
					"subject":   map[string]string{"reference": "Patient/" + patientID},
					"period":    map[string]string{"start": scheduledAt},
					"reasonCode": []map[string]interface{}{
						{"text": reason},
					},
					"participant": []map[string]interface{}{
						{
							"individual": map[string]string{"reference": "Practitioner/" + doctorID},
						},
					},
				},
				"request": map[string]string{"method": "PUT", "url": "Encounter/" + id},
			})
		}
	}

	// --- Prescriptions → FHIR MedicationRequest resources ---
	rxRows, err := db.Pool.Query(ctx, `
		SELECT id::text, medication_name, dosage, frequency, duration_days, status, created_at::text
		FROM prescriptions WHERE patient_id = $1 ORDER BY created_at DESC LIMIT 50
	`, claims.UserID)
	if err == nil {
		defer rxRows.Close()
		for rxRows.Next() {
			var id, med, dosage, freq, status, createdAt string
			var days int
			if err := rxRows.Scan(&id, &med, &dosage, &freq, &days, &status, &createdAt); err != nil {
				continue
			}
			fhirRxStatus := prescriptionStatusToFHIR(status)
			entries = append(entries, map[string]interface{}{
				"fullUrl": "urn:uuid:" + id,
				"resource": map[string]interface{}{
					"resourceType": "MedicationRequest",
					"id":           id,
					"status":       fhirRxStatus,
					"intent":       "order",
					"medicationCodeableConcept": map[string]interface{}{
						"text": med,
					},
					"subject": map[string]string{"reference": "Patient/" + patientID},
					"authoredOn": createdAt,
					"dosageInstruction": []map[string]interface{}{
						{
							"text": fmt.Sprintf("%s — %s for %d days", dosage, freq, days),
						},
					},
				},
				"request": map[string]string{"method": "PUT", "url": "MedicationRequest/" + id},
			})
		}
	}

	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"id":           "export-" + strings.ReplaceAll(patientID[:8], "-", ""),
		"type":         "transaction",
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"meta": map[string]interface{}{
			"profile": []string{"http://hl7.org/fhir/StructureDefinition/Bundle"},
		},
		"entry": entries,
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="healthonyx-fhir-export-%s.json"`, time.Now().Format("2006-01-02")))
	c.JSON(http.StatusOK, bundle)
}

func appointmentStatusToFHIR(status string) string {
	switch status {
	case "approved":
		return "planned"
	case "completed":
		return "finished"
	case "cancelled", "rejected":
		return "cancelled"
	default:
		return "unknown"
	}
}

func prescriptionStatusToFHIR(status string) string {
	switch status {
	case "active":
		return "active"
	case "completed":
		return "completed"
	case "revoked":
		return "stopped"
	default:
		return "unknown"
	}
}

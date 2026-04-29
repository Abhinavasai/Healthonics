package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/jackc/pgx/v5"
)

type FHIRBoundaryHandler struct{}

func NewFHIRBoundaryHandler() *FHIRBoundaryHandler {
	return &FHIRBoundaryHandler{}
}

type fhirBundle struct {
	ResourceType string      `json:"resourceType"`
	Type         string      `json:"type"`
	Entry        []fhirEntry `json:"entry"`
}

type fhirEntry struct {
	Resource map[string]any `json:"resource"`
}

type canonicalPatient struct {
	ID    string
	Email string
}

type canonicalEncounter struct {
	ID          string
	PatientID   string
	DoctorID    string
	Status      string
	StartedAt   time.Time
	Description string
}

type canonicalObservation struct {
	ID        string
	PatientID string
	Text      string
	CreatedAt time.Time
}

type canonicalMedicationRequest struct {
	ID           string
	PatientID    string
	DoctorID     string
	Status       string
	Medication   string
	Dosage       string
	Frequency    string
	CreatedAt    time.Time
	Instructions string
}

func validateFHIRBundleSchema(bundle fhirBundle) error {
	if strings.TrimSpace(bundle.ResourceType) != "Bundle" {
		return errBadRequest("resourceType must be Bundle")
	}
	if strings.TrimSpace(bundle.Type) == "" {
		return errBadRequest("bundle type is required")
	}
	for _, e := range bundle.Entry {
		rt, _ := e.Resource["resourceType"].(string)
		switch rt {
		case "Patient", "Encounter", "Observation", "MedicationRequest":
		default:
			return errBadRequest("unsupported FHIR resourceType in bundle")
		}
	}
	return nil
}

func errBadRequest(msg string) error { return &fhirValidationError{msg: msg} }

type fhirValidationError struct{ msg string }

func (e *fhirValidationError) Error() string { return e.msg }

func toFHIRPatient(p canonicalPatient) map[string]any {
	return map[string]any{
		"resourceType": "Patient",
		"id":           p.ID,
		"identifier": []map[string]any{
			{"system": "https://healthonyx.app/email", "value": p.Email},
		},
	}
}

func toFHIREncounter(e canonicalEncounter) map[string]any {
	return map[string]any{
		"resourceType": "Encounter",
		"id":           e.ID,
		"status":       mapEncounterStatus(e.Status),
		"subject":      map[string]any{"reference": "Patient/" + e.PatientID},
		"participant": []map[string]any{
			{"individual": map[string]any{"reference": "Practitioner/" + e.DoctorID}},
		},
		"period": map[string]any{
			"start": e.StartedAt.UTC().Format(time.RFC3339),
		},
		"reasonCode": []map[string]any{{"text": e.Description}},
	}
}

func toFHIRObservation(o canonicalObservation) map[string]any {
	return map[string]any{
		"resourceType": "Observation",
		"id":           o.ID,
		"status":       "final",
		"subject":      map[string]any{"reference": "Patient/" + o.PatientID},
		"issued":       o.CreatedAt.UTC().Format(time.RFC3339),
		"valueString":  o.Text,
	}
}

func toFHIRMedicationRequest(m canonicalMedicationRequest) map[string]any {
	return map[string]any{
		"resourceType": "MedicationRequest",
		"id":           m.ID,
		"status":       mapMedicationStatus(m.Status),
		"intent":       "order",
		"subject":      map[string]any{"reference": "Patient/" + m.PatientID},
		"requester":    map[string]any{"reference": "Practitioner/" + m.DoctorID},
		"medicationCodeableConcept": map[string]any{
			"text": m.Medication,
		},
		"dosageInstruction": []map[string]any{{
			"text": strings.TrimSpace(m.Dosage + " " + m.Frequency + " " + m.Instructions),
		}},
		"authoredOn": m.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func mapEncounterStatus(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "approved", "completed":
		return "finished"
	case "cancelled", "rejected":
		return "cancelled"
	default:
		return "planned"
	}
}

func mapMedicationStatus(s string) string {
	if strings.EqualFold(strings.TrimSpace(s), "revoked") {
		return "stopped"
	}
	return "active"
}

func parseBundleToCanonical(bundle fhirBundle) ([]canonicalPatient, []canonicalEncounter, []canonicalObservation, []canonicalMedicationRequest, error) {
	var patients []canonicalPatient
	var encounters []canonicalEncounter
	var observations []canonicalObservation
	var meds []canonicalMedicationRequest
	for _, e := range bundle.Entry {
		rt, _ := e.Resource["resourceType"].(string)
		switch rt {
		case "Patient":
			id, _ := e.Resource["id"].(string)
			identifier := ""
			if arr, ok := e.Resource["identifier"].([]any); ok && len(arr) > 0 {
				if first, ok := arr[0].(map[string]any); ok {
					identifier, _ = first["value"].(string)
				}
			}
			if strings.TrimSpace(id) == "" || strings.TrimSpace(identifier) == "" {
				return nil, nil, nil, nil, errBadRequest("patient must include id and identifier value")
			}
			patients = append(patients, canonicalPatient{ID: strings.TrimSpace(id), Email: strings.TrimSpace(strings.ToLower(identifier))})
		case "Encounter":
			id, _ := e.Resource["id"].(string)
			subj := readReferenceID(e.Resource, "subject")
			desc := readReasonText(e.Resource)
			if strings.TrimSpace(id) == "" || strings.TrimSpace(subj) == "" {
				return nil, nil, nil, nil, errBadRequest("encounter must include id and subject reference")
			}
			encounters = append(encounters, canonicalEncounter{
				ID: id, PatientID: subj, DoctorID: readParticipantReferenceID(e.Resource),
				Status: strings.TrimSpace(stringAny(e.Resource["status"])), StartedAt: time.Now().UTC(), Description: desc,
			})
		case "Observation":
			id, _ := e.Resource["id"].(string)
			subj := readReferenceID(e.Resource, "subject")
			txt := strings.TrimSpace(stringAny(e.Resource["valueString"]))
			if strings.TrimSpace(id) == "" || strings.TrimSpace(subj) == "" {
				return nil, nil, nil, nil, errBadRequest("observation must include id and subject reference")
			}
			observations = append(observations, canonicalObservation{ID: id, PatientID: subj, Text: txt, CreatedAt: time.Now().UTC()})
		case "MedicationRequest":
			id, _ := e.Resource["id"].(string)
			subj := readReferenceID(e.Resource, "subject")
			med := ""
			if medObj, ok := e.Resource["medicationCodeableConcept"].(map[string]any); ok {
				med = strings.TrimSpace(stringAny(medObj["text"]))
			}
			if strings.TrimSpace(id) == "" || strings.TrimSpace(subj) == "" || med == "" {
				return nil, nil, nil, nil, errBadRequest("medication request must include id, subject reference, and medication text")
			}
			meds = append(meds, canonicalMedicationRequest{
				ID: id, PatientID: subj, DoctorID: readReferenceID(e.Resource, "requester"),
				Status: strings.TrimSpace(stringAny(e.Resource["status"])), Medication: med, CreatedAt: time.Now().UTC(),
			})
		}
	}
	return patients, encounters, observations, meds, nil
}

func readReferenceID(resource map[string]any, field string) string {
	obj, ok := resource[field].(map[string]any)
	if !ok {
		return ""
	}
	ref := strings.TrimSpace(stringAny(obj["reference"]))
	parts := strings.Split(ref, "/")
	return strings.TrimSpace(parts[len(parts)-1])
}

func readParticipantReferenceID(resource map[string]any) string {
	arr, ok := resource["participant"].([]any)
	if !ok || len(arr) == 0 {
		return ""
	}
	first, ok := arr[0].(map[string]any)
	if !ok {
		return ""
	}
	ind, ok := first["individual"].(map[string]any)
	if !ok {
		return ""
	}
	ref := strings.TrimSpace(stringAny(ind["reference"]))
	parts := strings.Split(ref, "/")
	return strings.TrimSpace(parts[len(parts)-1])
}

func readReasonText(resource map[string]any) string {
	arr, ok := resource["reasonCode"].([]any)
	if !ok || len(arr) == 0 {
		return ""
	}
	first, ok := arr[0].(map[string]any)
	if !ok {
		return ""
	}
	return strings.TrimSpace(stringAny(first["text"]))
}

func stringAny(v any) string {
	s, _ := v.(string)
	return s
}

func (h *FHIRBoundaryHandler) ExportBundle(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	ctx := c.Request.Context()
	bundle := fhirBundle{ResourceType: "Bundle", Type: "collection", Entry: []fhirEntry{}}

	users, err := db.Pool.Query(ctx, `SELECT id::text, email FROM users WHERE role='patient' ORDER BY created_at DESC LIMIT 200`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	for users.Next() {
		var p canonicalPatient
		if scanErr := users.Scan(&p.ID, &p.Email); scanErr == nil {
			bundle.Entry = append(bundle.Entry, fhirEntry{Resource: toFHIRPatient(p)})
		}
	}
	users.Close()

	encRows, _ := db.Pool.Query(ctx, `
		SELECT id::text, patient_id::text, doctor_id::text, status, scheduled_at, reason
		FROM appointments ORDER BY created_at DESC LIMIT 200
	`)
	for encRows.Next() {
		var e canonicalEncounter
		if scanErr := encRows.Scan(&e.ID, &e.PatientID, &e.DoctorID, &e.Status, &e.StartedAt, &e.Description); scanErr == nil {
			bundle.Entry = append(bundle.Entry, fhirEntry{Resource: toFHIREncounter(e)})
		}
	}
	encRows.Close()

	obsRows, _ := db.Pool.Query(ctx, `
		SELECT id::text, patient_id::text, COALESCE(summary,''), created_at
		FROM patient_documents ORDER BY created_at DESC LIMIT 200
	`)
	for obsRows.Next() {
		var o canonicalObservation
		if scanErr := obsRows.Scan(&o.ID, &o.PatientID, &o.Text, &o.CreatedAt); scanErr == nil {
			bundle.Entry = append(bundle.Entry, fhirEntry{Resource: toFHIRObservation(o)})
		}
	}
	obsRows.Close()

	medRows, _ := db.Pool.Query(ctx, `
		SELECT id::text, patient_id::text, doctor_id::text, status, medication_name, dosage, frequency, created_at, instructions
		FROM prescriptions ORDER BY created_at DESC LIMIT 200
	`)
	for medRows.Next() {
		var m canonicalMedicationRequest
		if scanErr := medRows.Scan(&m.ID, &m.PatientID, &m.DoctorID, &m.Status, &m.Medication, &m.Dosage, &m.Frequency, &m.CreatedAt, &m.Instructions); scanErr == nil {
			bundle.Entry = append(bundle.Entry, fhirEntry{Resource: toFHIRMedicationRequest(m)})
		}
	}
	medRows.Close()

	c.JSON(http.StatusOK, gin.H{"bundle": bundle})
}

func (h *FHIRBoundaryHandler) ImportBundle(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	var payload struct {
		Bundle fhirBundle `json:"bundle"`
		Apply  bool       `json:"apply"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := validateFHIRBundleSchema(payload.Bundle); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	patients, encounters, observations, meds, err := parseBundleToCanonical(payload.Bundle)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if payload.Apply {
		if err := applyFHIRImport(c, patients, encounters, observations, meds); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"validated":          true,
		"applied":            payload.Apply,
		"patients_count":     len(patients),
		"encounters_count":   len(encounters),
		"observations_count": len(observations),
		"medications_count":  len(meds),
	})
}

func applyFHIRImport(c *gin.Context, patients []canonicalPatient, encounters []canonicalEncounter, observations []canonicalObservation, meds []canonicalMedicationRequest) error {
	ctx := c.Request.Context()
	for _, p := range patients {
		_, _ = db.Pool.Exec(ctx, `UPDATE users SET email=$2 WHERE id::text=$1 AND role='patient'`, p.ID, p.Email)
	}
	for _, e := range encounters {
		_, _ = db.Pool.Exec(ctx, `UPDATE appointments SET reason=$2, status=$3 WHERE id::text=$1`, e.ID, strings.TrimSpace(e.Description), strings.TrimSpace(strings.ToLower(e.Status)))
	}
	for _, o := range observations {
		if _, err := uuid.Parse(strings.TrimSpace(o.PatientID)); err != nil {
			continue
		}
		_, _ = db.Pool.Exec(ctx, `
			INSERT INTO patient_documents (patient_id, filename, content_type, size_bytes, status, body, summary, summary_status)
			VALUES ($1::uuid, $2, 'text/plain', $3, 'ready', $4::bytea, $5, 'ready')
			ON CONFLICT DO NOTHING
		`, o.PatientID, "fhir-observation-"+o.ID+".txt", len(o.Text), []byte(o.Text), o.Text)
	}
	for _, m := range meds {
		if _, err := db.Pool.Exec(ctx, `
			UPDATE prescriptions
			SET medication_name=$2,
				dosage=COALESCE(NULLIF($3,''), dosage),
				frequency=COALESCE(NULLIF($4,''), frequency),
				instructions=COALESCE(NULLIF($5,''), instructions),
				status=$6
			WHERE id::text=$1
		`, m.ID, m.Medication, m.Dosage, m.Frequency, m.Instructions, normalizeMedicationStatusForDB(m.Status)); err != nil && err != pgx.ErrNoRows {
			return err
		}
	}
	return nil
}

func normalizeMedicationStatusForDB(s string) string {
	if strings.EqualFold(strings.TrimSpace(s), "stopped") {
		return "revoked"
	}
	return "active"
}

func bundleToJSON(bundle fhirBundle) []byte {
	b, _ := json.Marshal(bundle)
	return b
}

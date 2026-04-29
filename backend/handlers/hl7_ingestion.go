package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
)

type HL7LabIngestionHandler struct{}

func NewHL7LabIngestionHandler() *HL7LabIngestionHandler {
	return &HL7LabIngestionHandler{}
}

type hl7LabObservation struct {
	Code  string
	Value string
	Units string
}

type hl7LabMessage struct {
	MessageControlID  string
	PatientIdentifier string
	Observations      []hl7LabObservation
}

func parseHL7LabMessage(raw string) (hl7LabMessage, error) {
	lines := splitHL7Segments(raw)
	msg := hl7LabMessage{}
	var hasMSH, hasPID bool
	for _, line := range lines {
		fields := strings.Split(line, "|")
		if len(fields) == 0 {
			continue
		}
		switch strings.TrimSpace(fields[0]) {
		case "MSH":
			hasMSH = true
			if len(fields) > 9 {
				msg.MessageControlID = strings.TrimSpace(fields[9])
			}
		case "PID":
			hasPID = true
			if len(fields) > 3 {
				msg.PatientIdentifier = strings.TrimSpace(firstHL7Component(fields[3]))
			}
		case "OBX":
			if len(fields) < 6 {
				continue
			}
			msg.Observations = append(msg.Observations, hl7LabObservation{
				Code:  strings.TrimSpace(firstHL7Component(fields[3])),
				Value: strings.TrimSpace(fields[5]),
				Units: strings.TrimSpace(fields[6]),
			})
		}
	}
	if !hasMSH {
		return hl7LabMessage{}, errBadRequest("missing MSH segment")
	}
	if !hasPID {
		return hl7LabMessage{}, errBadRequest("missing PID segment")
	}
	if strings.TrimSpace(msg.PatientIdentifier) == "" {
		return hl7LabMessage{}, errBadRequest("missing patient identifier in PID-3")
	}
	if len(msg.Observations) == 0 {
		return hl7LabMessage{}, errBadRequest("missing OBX lab observations")
	}
	return msg, nil
}

func splitHL7Segments(raw string) []string {
	norm := strings.ReplaceAll(raw, "\r\n", "\n")
	norm = strings.ReplaceAll(norm, "\r", "\n")
	parts := strings.Split(norm, "\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func firstHL7Component(v string) string {
	idx := strings.Index(v, "^")
	if idx == -1 {
		return v
	}
	return v[:idx]
}

func reconcilePatientIdentifier(c *gin.Context, patientIdentifier string) (string, bool) {
	var patientID string
	err := db.Pool.QueryRow(c.Request.Context(), `
		SELECT id::text
		FROM users
		WHERE role = 'patient' AND (id::text = $1 OR LOWER(email) = LOWER($1))
		LIMIT 1
	`, strings.TrimSpace(patientIdentifier)).Scan(&patientID)
	if err != nil {
		return "", false
	}
	return patientID, true
}

func (h *HL7LabIngestionHandler) IngestLabResult(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	var req struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	parsed, err := parseHL7LabMessage(req.Message)
	if err != nil {
		_, _ = db.Pool.Exec(c.Request.Context(), `
			INSERT INTO hl7_lab_ingestion_events (
				message_control_id, patient_identifier, transform_status, transform_outcome, hl7_message
			) VALUES ($1, $2, 'rejected_malformed', $3, $4)
		`, "", "", err.Error(), req.Message)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	patientID, matched := reconcilePatientIdentifier(c, parsed.PatientIdentifier)
	outcomes := make([]gin.H, 0, len(parsed.Observations))
	for _, ob := range parsed.Observations {
		status := "unmatched_patient"
		outcome := "parsed_ok_unmatched_patient"
		var pid any = nil
		if matched {
			status = "reconciled"
			outcome = "parsed_ok_patient_reconciled"
			if parsedID, parseErr := uuid.Parse(patientID); parseErr == nil {
				pid = parsedID
			}
		}
		_, _ = db.Pool.Exec(c.Request.Context(), `
			INSERT INTO hl7_lab_ingestion_events (
				message_control_id, patient_identifier, patient_id, observation_code, observation_value, units,
				transform_status, transform_outcome, hl7_message
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, parsed.MessageControlID, parsed.PatientIdentifier, pid, ob.Code, ob.Value, ob.Units, status, outcome, req.Message)
		outcomes = append(outcomes, gin.H{
			"observation_code": ob.Code,
			"status":           status,
			"outcome":          outcome,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message_control_id": parsed.MessageControlID,
		"patient_identifier": parsed.PatientIdentifier,
		"observations_count": len(parsed.Observations),
		"patient_reconciled": matched,
		"outcomes":           outcomes,
	})
}

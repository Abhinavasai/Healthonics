package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	AppointmentStatusPending             = "pending"
	AppointmentStatusApproved            = "approved"
	AppointmentStatusRejected            = "rejected"
	AppointmentStatusCancelled           = "cancelled"
	AppointmentStatusRescheduleRequested = "reschedule_requested"
	AppointmentStatusCompleted           = "completed"
	AppointmentStatusNoShow              = "no_show"
)

type Appointment struct {
	ID                 uuid.UUID  `json:"id"`
	PatientID          uuid.UUID  `json:"patient_id"`
	DoctorID           uuid.UUID  `json:"doctor_id"`
	ScheduledAt        time.Time  `json:"scheduled_at"`
	Reason             string     `json:"reason"`
	Status             string     `json:"status"`
	PendingScheduledAt *time.Time `json:"pending_scheduled_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	AppointmentStatusPending            = "pending"
	AppointmentStatusApproved         = "approved"
	AppointmentStatusRejected         = "rejected"
	AppointmentStatusCancelled        = "cancelled"
	AppointmentStatusCompleted        = "completed"
	AppointmentStatusNoShow           = "no_show"
	AppointmentStatusRescheduleRequested = "reschedule_requested"
)

type Appointment struct {
	ID           uuid.UUID  `json:"id"`
	PatientID    uuid.UUID  `json:"patient_id"`
	PatientEmail string     `json:"patient_email,omitempty"`
	DoctorID     uuid.UUID  `json:"doctor_id"`
	DoctorEmail  string     `json:"doctor_email,omitempty"`
	ScheduledAt  time.Time  `json:"scheduled_at"`
	Reason       string     `json:"reason"`
	Status       string     `json:"status"`
	VideoLink    *string    `json:"video_link,omitempty"`
	CheckedInAt  *time.Time `json:"checked_in_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

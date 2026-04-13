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
	ID          uuid.UUID `json:"id"`
	PatientID   uuid.UUID `json:"patient_id"`
	DoctorID    uuid.UUID `json:"doctor_id"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Reason      string    `json:"reason"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

package models

import (
	"time"

	"github.com/google/uuid"
)

type AppointmentActivity struct {
	ID            uuid.UUID `json:"id"`
	AppointmentID uuid.UUID `json:"appointment_id"`
	ActorUserID   uuid.UUID `json:"actor_user_id"`
	ActorEmail    string    `json:"actor_email,omitempty"`
	Action        string    `json:"action"`
	Detail        string    `json:"detail,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type immutableAuditEvent struct {
	Action     string
	EntityType string
	EntityID   string
	Detail     string
	EventTime  time.Time
	PrevHash   string
	EventHash  string
}

func computeImmutableAuditHash(prevHash, action, entityType, entityID, detail string, eventTime time.Time) string {
	// Delimiter-separated canonical payload for deterministic checksum generation.
	payload := strings.Join([]string{
		strings.TrimSpace(prevHash),
		strings.TrimSpace(action),
		strings.TrimSpace(entityType),
		strings.TrimSpace(entityID),
		strings.TrimSpace(detail),
		eventTime.UTC().Format(time.RFC3339Nano),
	}, "|")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func verifyImmutableAuditChain(events []immutableAuditEvent) error {
	prev := ""
	for i, e := range events {
		expected := computeImmutableAuditHash(prev, e.Action, e.EntityType, e.EntityID, e.Detail, e.EventTime)
		if e.EventHash != expected {
			return fmt.Errorf("immutable audit checksum mismatch at index %d", i)
		}
		if e.PrevHash != prev {
			return fmt.Errorf("immutable audit prev_hash mismatch at index %d", i)
		}
		prev = e.EventHash
	}
	return nil
}

func writeAudit(ctx context.Context, tx pgx.Tx, actor uuid.UUID, action, entityType, entityID, detail string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id, detail)
		VALUES ($1, $2, $3, $4, $5)
	`, actor, action, entityType, entityID, detail)
	if err != nil {
		return err
	}

	var prevHash string
	_ = tx.QueryRow(ctx, `
		SELECT COALESCE(event_hash, '')
		FROM immutable_audit_events
		ORDER BY sequence DESC
		LIMIT 1
	`).Scan(&prevHash)

	eventTime := time.Now().UTC()
	eventHash := computeImmutableAuditHash(prevHash, action, entityType, entityID, detail, eventTime)
	_, err = tx.Exec(ctx, `
		INSERT INTO immutable_audit_events (
			actor_user_id, action, entity_type, entity_id, detail, event_time, prev_hash, event_hash
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, actor, action, entityType, entityID, detail, eventTime, prevHash, eventHash)
	return err
}

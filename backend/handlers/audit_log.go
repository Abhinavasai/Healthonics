package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func writeAudit(ctx context.Context, tx pgx.Tx, actor uuid.UUID, action, entityType, entityID, detail string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id, detail)
		VALUES ($1, $2, $3, $4, $5)
	`, actor, action, entityType, entityID, detail)
	return err
}

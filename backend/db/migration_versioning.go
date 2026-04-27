package db

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Migration is a single ordered schema/data change unit.
// ID should be sortable (for example: "202604271500_add_indexes").
type Migration struct {
	ID          string
	Description string
	Apply       func(context.Context) error
}

// VersionStore tracks which migrations were applied.
type VersionStore interface {
	Ensure(ctx context.Context) error
	IsApplied(ctx context.Context, migrationID string) (bool, error)
	MarkApplied(ctx context.Context, migration Migration) error
}

// SQLVersionStore persists migration versions in Postgres.
type SQLVersionStore struct{}

func (s SQLVersionStore) Ensure(ctx context.Context) error {
	if Pool == nil {
		return errors.New("db pool not initialized")
	}
	_, err := Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			migration_id TEXT PRIMARY KEY,
			description  TEXT NOT NULL DEFAULT '',
			applied_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func (s SQLVersionStore) IsApplied(ctx context.Context, migrationID string) (bool, error) {
	if Pool == nil {
		return false, errors.New("db pool not initialized")
	}
	var found bool
	err := Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE migration_id = $1)
	`, migrationID).Scan(&found)
	return found, err
}

func (s SQLVersionStore) MarkApplied(ctx context.Context, migration Migration) error {
	if Pool == nil {
		return errors.New("db pool not initialized")
	}
	_, err := Pool.Exec(ctx, `
		INSERT INTO schema_migrations (migration_id, description)
		VALUES ($1, $2)
		ON CONFLICT (migration_id) DO NOTHING
	`, migration.ID, migration.Description)
	return err
}

// ApplyPendingMigrations runs all unapplied migrations in deterministic order.
// It returns the IDs that were applied during this run.
func ApplyPendingMigrations(ctx context.Context, store VersionStore, migrations []Migration) ([]string, error) {
	if store == nil {
		return nil, errors.New("version store is required")
	}
	if err := store.Ensure(ctx); err != nil {
		return nil, fmt.Errorf("ensure version store: %w", err)
	}

	ordered, err := normalizeMigrations(migrations)
	if err != nil {
		return nil, err
	}

	appliedNow := make([]string, 0, len(ordered))
	for _, m := range ordered {
		alreadyApplied, err := store.IsApplied(ctx, m.ID)
		if err != nil {
			return nil, fmt.Errorf("check migration %s: %w", m.ID, err)
		}
		if alreadyApplied {
			continue
		}
		if err := m.Apply(ctx); err != nil {
			return appliedNow, fmt.Errorf("apply migration %s: %w", m.ID, err)
		}
		if err := store.MarkApplied(ctx, m); err != nil {
			return appliedNow, fmt.Errorf("mark migration %s: %w", m.ID, err)
		}
		appliedNow = append(appliedNow, m.ID)
	}
	return appliedNow, nil
}

func normalizeMigrations(migrations []Migration) ([]Migration, error) {
	out := make([]Migration, len(migrations))
	copy(out, migrations)
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})

	seen := map[string]struct{}{}
	for _, m := range out {
		if strings.TrimSpace(m.ID) == "" {
			return nil, errors.New("migration id is required")
		}
		if m.Apply == nil {
			return nil, fmt.Errorf("migration %s has nil apply function", m.ID)
		}
		if _, exists := seen[m.ID]; exists {
			return nil, fmt.Errorf("duplicate migration id: %s", m.ID)
		}
		seen[m.ID] = struct{}{}
	}
	return out, nil
}

package db

import (
	"context"
	"errors"
	"slices"
	"testing"
)

type memoryVersionStore struct {
	applied map[string]Migration
}

func (m *memoryVersionStore) Ensure(context.Context) error {
	if m.applied == nil {
		m.applied = map[string]Migration{}
	}
	return nil
}

func (m *memoryVersionStore) IsApplied(_ context.Context, migrationID string) (bool, error) {
	_, ok := m.applied[migrationID]
	return ok, nil
}

func (m *memoryVersionStore) MarkApplied(_ context.Context, migration Migration) error {
	m.applied[migration.ID] = migration
	return nil
}

func TestApplyPendingMigrations_SortsAndAppliesOnlyPending(t *testing.T) {
	store := &memoryVersionStore{
		applied: map[string]Migration{
			"202604270100_existing": {ID: "202604270100_existing"},
		},
	}
	var applyOrder []string
	migrations := []Migration{
		{
			ID: "202604270300_third",
			Apply: func(context.Context) error {
				applyOrder = append(applyOrder, "202604270300_third")
				return nil
			},
		},
		{
			ID: "202604270100_existing",
			Apply: func(context.Context) error {
				applyOrder = append(applyOrder, "202604270100_existing")
				return nil
			},
		},
		{
			ID: "202604270200_second",
			Apply: func(context.Context) error {
				applyOrder = append(applyOrder, "202604270200_second")
				return nil
			},
		},
	}

	applied, err := ApplyPendingMigrations(context.Background(), store, migrations)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantApplyOrder := []string{"202604270200_second", "202604270300_third"}
	if !slices.Equal(applyOrder, wantApplyOrder) {
		t.Fatalf("apply order mismatch: got %v want %v", applyOrder, wantApplyOrder)
	}
	if !slices.Equal(applied, wantApplyOrder) {
		t.Fatalf("applied result mismatch: got %v want %v", applied, wantApplyOrder)
	}
}

func TestApplyPendingMigrations_StopsOnApplyError(t *testing.T) {
	store := &memoryVersionStore{}
	migrations := []Migration{
		{
			ID: "202604270100_first",
			Apply: func(context.Context) error {
				return nil
			},
		},
		{
			ID: "202604270200_failing",
			Apply: func(context.Context) error {
				return errors.New("boom")
			},
		},
		{
			ID: "202604270300_never_runs",
			Apply: func(context.Context) error {
				t.Fatal("expected third migration not to run")
				return nil
			},
		},
	}

	applied, err := ApplyPendingMigrations(context.Background(), store, migrations)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(applied) != 1 || applied[0] != "202604270100_first" {
		t.Fatalf("unexpected applied list before failure: %v", applied)
	}
	if _, ok := store.applied["202604270100_first"]; !ok {
		t.Fatal("expected successful migration to be marked applied")
	}
	if _, ok := store.applied["202604270200_failing"]; ok {
		t.Fatal("did not expect failing migration to be marked applied")
	}
}

func TestApplyPendingMigrations_ValidatesDuplicateIDs(t *testing.T) {
	store := &memoryVersionStore{}
	migrations := []Migration{
		{
			ID: "202604270100_dup",
			Apply: func(context.Context) error {
				return nil
			},
		},
		{
			ID: "202604270100_dup",
			Apply: func(context.Context) error {
				return nil
			},
		},
	}

	if _, err := ApplyPendingMigrations(context.Background(), store, migrations); err == nil {
		t.Fatal("expected duplicate id error")
	}
}

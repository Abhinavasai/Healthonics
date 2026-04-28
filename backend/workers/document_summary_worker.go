package workers

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/healthonyx/backend/handlers"
	"github.com/jackc/pgx/v5"
)

type DocumentSummaryWorker struct {
	interval time.Duration
	process  func(context.Context) (int64, error)
}

func NewDocumentSummaryWorker(interval time.Duration) *DocumentSummaryWorker {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	return &DocumentSummaryWorker{interval: interval}
}

func (w *DocumentSummaryWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	if _, err := w.processPending(ctx); err != nil {
		log.Printf("document summary worker initial process error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := w.processPending(ctx); err != nil {
				log.Printf("document summary worker process error: %v", err)
			}
		}
	}
}

func (w *DocumentSummaryWorker) processPending(ctx context.Context) (int64, error) {
	if w.process != nil {
		return w.process(ctx)
	}
	return w.ProcessPending(ctx)
}

func (w *DocumentSummaryWorker) ProcessPending(ctx context.Context) (int64, error) {
	var processed int64
	for {
		ok, err := processSingleSummaryJob(ctx)
		if err != nil {
			return processed, err
		}
		if !ok {
			return processed, nil
		}
		processed++
	}
}

func processSingleSummaryJob(ctx context.Context) (bool, error) {
	var jobID, docID uuid.UUID
	err := db.Pool.QueryRow(ctx, `
		WITH next_job AS (
			SELECT id, document_id
			FROM document_summary_jobs
			WHERE status = 'pending'
			  AND run_after <= NOW()
			ORDER BY created_at
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE document_summary_jobs j
		SET status = 'processing',
			started_at = NOW(),
			finished_at = NULL,
			last_error = NULL,
			attempts = j.attempts + 1
		FROM next_job
		WHERE j.id = next_job.id
		RETURNING j.id, j.document_id
	`).Scan(&jobID, &docID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	var filename, contentType string
	var sizeBytes int64
	var body []byte
	err = db.Pool.QueryRow(ctx, `
		SELECT filename, content_type, size_bytes, body
		FROM patient_documents
		WHERE id = $1
	`, docID).Scan(&filename, &contentType, &sizeBytes, &body)
	if err != nil {
		if err == pgx.ErrNoRows {
			_, _ = db.Pool.Exec(ctx, `
				UPDATE document_summary_jobs
				SET status = 'failed', finished_at = NOW(), last_error = 'document not found'
				WHERE id = $1
			`, jobID)
			return true, nil
		}
		return false, err
	}

	extracted, exErr := handlers.ExtractDocumentTextForWorker(contentType, filename, body)
	if exErr != nil {
		msg := handlers.NormalizeSummaryExtractionErrorForWorker(exErr)
		_, _ = db.Pool.Exec(ctx, `
			UPDATE patient_documents
			SET summary_status = 'failed', summary_error = $2
			WHERE id = $1
		`, docID, msg)
		_, _ = db.Pool.Exec(ctx, `
			UPDATE document_summary_jobs
			SET status = 'failed', finished_at = NOW(), last_error = $2
			WHERE id = $1
		`, jobID, msg)
		return true, nil
	}

	settings, settingsErr := handlers.LoadAIRuntimeSettingsForWorker(ctx)
	if settingsErr != nil {
		_, _ = db.Pool.Exec(ctx, `
			UPDATE patient_documents
			SET summary_status = 'failed', summary_error = $2
			WHERE id = $1
		`, docID, settingsErr.Error())
		_, _ = db.Pool.Exec(ctx, `
			UPDATE document_summary_jobs
			SET status = 'failed', finished_at = NOW(), last_error = $2
			WHERE id = $1
		`, jobID, settingsErr.Error())
		return true, nil
	}

	summary, mode, summarizeErr := handlers.GenerateSummaryWithAIRuntime(ctx, settings, filename, sizeBytes, extracted)
	if summarizeErr != nil {
		_, _ = db.Pool.Exec(ctx, `
			UPDATE patient_documents
			SET summary_status = 'failed', summary_error = $2
			WHERE id = $1
		`, docID, summarizeErr.Error())
		_, _ = db.Pool.Exec(ctx, `
			UPDATE document_summary_jobs
			SET status = 'failed', finished_at = NOW(), last_error = $2
			WHERE id = $1
		`, jobID, summarizeErr.Error())
		return true, nil
	}
	if mode != "" && mode != "ollama" {
		summary = summary + "\n\n[summary_mode=" + mode + "]"
	}

	_, err = db.Pool.Exec(ctx, `
		UPDATE patient_documents
		SET summary = $2, summary_status = 'ready', summary_error = NULL
		WHERE id = $1
	`, docID, summary)
	if err != nil {
		_, _ = db.Pool.Exec(ctx, `
			UPDATE patient_documents
			SET summary_status = 'failed', summary_error = $2
			WHERE id = $1
		`, docID, err.Error())
		_, _ = db.Pool.Exec(ctx, `
			UPDATE document_summary_jobs
			SET status = 'failed', finished_at = NOW(), last_error = $2
			WHERE id = $1
		`, jobID, err.Error())
		return true, nil
	}

	_, err = db.Pool.Exec(ctx, `
		UPDATE document_summary_jobs
		SET status = 'done', finished_at = NOW(), last_error = NULL
		WHERE id = $1
	`, jobID)
	if err != nil {
		return false, err
	}
	return true, nil
}

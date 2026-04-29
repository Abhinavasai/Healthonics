package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/config"
	"github.com/healthonyx/backend/db"
	"github.com/healthonyx/backend/handlers"
	"github.com/healthonyx/backend/workers"
)

type soakEvidence struct {
	SoakDurationMinutes float64  `json:"soak_duration_minutes"`
	RequestsTotal       int      `json:"requests_total"`
	ErrorsTotal         int      `json:"errors_total"`
	ErrorRatePercent    float64  `json:"error_rate_percent"`
	P95LatencyMS        int      `json:"p95_latency_ms"`
	AvgLatencyMS        int      `json:"avg_latency_ms"`
	PathSample          []string `json:"path_sample,omitempty"`
}

type chaosScenarioResult struct {
	Name    string                 `json:"name"`
	Passed  bool                   `json:"passed"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type reliabilityEvidence struct {
	ChaosScenariosTotal  int                   `json:"chaos_scenarios_total"`
	ChaosScenariosPassed int                   `json:"chaos_scenarios_passed"`
	Suite                string                `json:"suite,omitempty"`
	Scenarios            []chaosScenarioResult `json:"scenarios"`
}

func main() {
	mode := flag.String("mode", "", "soak|chaos")
	output := flag.String("output", "", "output path for evidence json")

	// Soak parameters
	baseURL := flag.String("base-url", "http://127.0.0.1:8080", "backend base URL")
	patientEmail := flag.String("patient-email", "patient@healthonyx.demo", "patient email")
	patientPassword := flag.String("patient-password", "patient123", "patient password")
	doctorEmail := flag.String("doctor-email", "doctor@healthonyx.demo", "doctor email")
	doctorPassword := flag.String("doctor-password", "doctor123", "doctor password")
	durationSeconds := flag.Int("duration-seconds", 300, "soak duration seconds (gate requires >= 300s)")
	concurrency := flag.Int("concurrency", 8, "total concurrent workers")

	// Chaos parameters
	sendgridWebhookSecret := flag.String("sendgrid-webhook-secret", "demo-sendgrid-secret", "sendgrid webhook secret used for signature verification")
	notificationRecipientEmail := flag.String("chaos-notification-recipient-email", "patient@healthonyx.demo", "recipient email for seeded notifications")
	seedMigrations := flag.Bool("migrate", true, "run db.Migrate() before chaos scenarios")

	flag.Parse()

	if *mode == "" {
		log.Fatal("missing --mode (soak|chaos)")
	}
	if *output == "" {
		log.Fatal("missing --output")
	}

	ctx := context.Background()
	cfg := config.Load()
	if cfg.DatabaseURL == "" && strings.EqualFold(strings.TrimSpace(*mode), "chaos") {
		log.Fatal("DATABASE_URL is required for chaos evidence generation")
	}

	switch strings.ToLower(strings.TrimSpace(*mode)) {
	case "soak":
		if err := runSoak(ctx, *baseURL, *patientEmail, *patientPassword, *doctorEmail, *doctorPassword, time.Duration(*durationSeconds)*time.Second, *concurrency, *output); err != nil {
			log.Fatal(err)
		}
	case "chaos":
		if err := runChaos(ctx, cfg.DatabaseURL, *seedMigrations, *sendgridWebhookSecret, *notificationRecipientEmail, *output); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown --mode: %s", *mode)
	}
}

func runSoak(ctx context.Context, baseURL, patientEmail, patientPassword, doctorEmail, doctorPassword string, duration time.Duration, concurrency int, outputPath string) error {
	client := &http.Client{Timeout: 15 * time.Second}

	patientToken, err := loginAndGetBearer(ctx, client, baseURL, patientEmail, patientPassword)
	if err != nil {
		return fmt.Errorf("patient login: %w", err)
	}
	doctorToken, err := loginAndGetBearer(ctx, client, baseURL, doctorEmail, doctorPassword)
	if err != nil {
		return fmt.Errorf("doctor login: %w", err)
	}

	type requestDef struct {
		name  string
		path  string
		token string
	}
	requests := []requestDef{
		{name: "patient-dashboard-summary", path: "/api/patient/dashboard/summary", token: patientToken},
		{name: "doctor-dashboard-summary", path: "/api/doctor/dashboard/summary", token: doctorToken},
		{name: "health", path: "/health", token: ""},
	}

	samplePaths := make([]string, 0, len(requests))
	for _, r := range requests {
		samplePaths = append(samplePaths, r.path)
	}

	deadline := time.Now().Add(duration)
	latencies := make([]int, 0, 1024)

	var mu sync.Mutex
	errorsTotal := 0
	requestsTotal := 0

	worker := func(wi int) {
		idx := wi % len(requests)
		for time.Now().Before(deadline) {
			r := requests[idx]
			idx = (idx + 1) % len(requests)

			start := time.Now()
			ok := false

			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+r.path, nil)
			req.Header.Set("Accept", "application/json")
			if strings.TrimSpace(r.token) != "" {
				req.Header.Set("Authorization", "Bearer "+r.token)
			}

			resp, reqErr := client.Do(req)
			if reqErr == nil && resp != nil {
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				ok = resp.StatusCode >= 200 && resp.StatusCode < 300
			}
			latencyMS := int(time.Since(start).Milliseconds())

			mu.Lock()
			requestsTotal++
			if !ok {
				errorsTotal++
			}
			latencies = append(latencies, latencyMS)
			mu.Unlock()
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			worker(i)
		}(i)
	}
	wg.Wait()

	p95 := percentile95(latencies)
	avg := averageInt(latencies)
	errorRate := 0.0
	if requestsTotal > 0 {
		errorRate = (float64(errorsTotal) / float64(requestsTotal)) * 100
	}

	ev := soakEvidence{
		SoakDurationMinutes: duration.Minutes(),
		RequestsTotal:       requestsTotal,
		ErrorsTotal:         errorsTotal,
		ErrorRatePercent:    round2(errorRate),
		P95LatencyMS:        p95,
		AvgLatencyMS:        avg,
		PathSample:          samplePaths,
	}

	if err := writeJSONFile(outputPath, ev); err != nil {
		return err
	}
	fmt.Printf("soak evidence written: %s\n", outputPath)
	return nil
}

func loginAndGetBearer(ctx context.Context, client *http.Client, baseURL, email, password string) (string, error) {
	loginURL := strings.TrimRight(baseURL, "/") + "/api/login"
	payload := map[string]string{"email": email, "password": password}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _, _ = io.Copy(io.Discard, resp.Body); _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed status=%d body=%s", resp.StatusCode, string(raw))
	}

	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if strings.TrimSpace(out.Token) == "" {
		return "", fmt.Errorf("login returned empty token")
	}
	return out.Token, nil
}

func runChaos(ctx context.Context, databaseURL string, migrate bool, sendgridWebhookSecret, recipientEmail, outputPath string) error {
	if err := db.Connect(databaseURL); err != nil {
		return err
	}
	defer db.Close()

	if migrate {
		if err := db.Migrate(ctx); err != nil {
			return err
		}
	}

	scenarios := []chaosScenarioResult{
		{Name: "notification-worker-provider-outage", Passed: false, Details: map[string]interface{}{}},
		{Name: "provider-callback-replay-dedup", Passed: false, Details: map[string]interface{}{}},
	}

	// Scenario 1: notification worker with nil providers should dead-letter with NTF_PROVIDER_CONFIG_MISSING.
	{
		notificationID, err := seedEmailNotificationForChaos(ctx, recipientEmail)
		if err != nil {
			return fmt.Errorf("seed scenario1 notification: %w", err)
		}

		worker := workers.NewNotificationWorkerWithConfig(workers.NotificationWorkerConfig{
			Interval:   1 * time.Second,
			MaxRetries: 1,
			Providers:  workers.NewNotificationProviderSet(nil, nil),
		})

		processed, err := worker.ProcessDue(ctx)
		if err != nil {
			return fmt.Errorf("worker ProcessDue (scenario1): %w", err)
		}

		deadLetters, deliveredAttempts, finalErr, err := chaosAssertNotificationDeadLetter(ctx, notificationID)
		if err != nil {
			return fmt.Errorf("assert scenario1: %w", err)
		}

		passed := processed > 0 &&
			deadLetters == 1 &&
			deliveredAttempts >= 1 &&
			strings.Contains(finalErr, "NTF_PROVIDER_CONFIG_MISSING")

		scenarios[0].Passed = passed
		scenarios[0].Details = map[string]interface{}{
			"processed_due":           processed,
			"dead_letters_count":      deadLetters,
			"delivery_attempts_count": deliveredAttempts,
			"final_error_contains":    "NTF_PROVIDER_CONFIG_MISSING",
			"final_error_preview":     truncateString(finalErr, 240),
		}
	}

	// Scenario 2: provider callback replay should not duplicate callback events or re-apply reconciliation.
	{
		notificationID, err := seedEmailNotificationForChaos(ctx, recipientEmail)
		if err != nil {
			return fmt.Errorf("seed scenario2 notification: %w", err)
		}

		eventID := "SG-DEDUP-" + time.Now().UTC().Format("20060102T150405") + "-" + notificationID.String()
		patientEmail := strings.TrimSpace(strings.ToLower(recipientEmail))

		payloadBytes, payloadBody, err := buildSendGridBouncePayload(eventID, notificationID, patientEmail)
		if err != nil {
			return fmt.Errorf("build sendgrid payload: %w", err)
		}

		// Provider callback endpoints are not JWT-protected; we can call handlers directly using an in-process router.
		engine := gin.New()
		cbHandler := handlers.NewProviderCallbacksHandler(sendgridWebhookSecret, "")
		engine.POST("/api/provider-callbacks/sendgrid", cbHandler.SendGridWebhook)

		ts := httptest.NewServer(engine)
		defer ts.Close()

		// Handler signature verification uses the raw request body and the timestamp header.
		nowTS := fmt.Sprintf("%d", time.Now().UTC().Unix())
		sig := computeSendGridSignature(sendgridWebhookSecret, nowTS, payloadBody)

		for i := 0; i < 2; i++ {
			req, _ := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+"/api/provider-callbacks/sendgrid", bytes.NewReader(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
			req.Header.Set("X-Twilio-Email-Event-Webhook-Timestamp", nowTS)
			req.Header.Set("X-Twilio-Email-Event-Webhook-Signature", sig)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("sendgrid webhook call %d: %w", i+1, err)
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}

		eventsCount, finalErr, attemptCount, notifStatus, err := chaosAssertSendGridReplay(ctx, notificationID, eventID)
		if err != nil {
			return fmt.Errorf("assert scenario2: %w", err)
		}

		passed := eventsCount == 1 &&
			strings.Contains(notifStatus, "failed") &&
			attemptCount == 1 &&
			strings.Contains(finalErr, "NTF_CALLBACK_HARD_FAILURE")

		scenarios[1].Passed = passed
		scenarios[1].Details = map[string]interface{}{
			"provider_callback_events_count":         eventsCount,
			"notification_status":                    notifStatus,
			"notification_dead_letter_attempt_count": attemptCount,
			"final_error_contains":                   "NTF_CALLBACK_HARD_FAILURE",
			"final_error_preview":                    truncateString(finalErr, 240),
		}
	}

	passedCount := 0
	for _, s := range scenarios {
		if s.Passed {
			passedCount++
		}
	}

	ev := reliabilityEvidence{
		ChaosScenariosTotal:  len(scenarios),
		ChaosScenariosPassed: passedCount,
		Suite:                "provider-outage-and-callback-replay",
		Scenarios:            scenarios,
	}

	if err := writeJSONFile(outputPath, ev); err != nil {
		return err
	}
	fmt.Printf("chaos evidence written: %s\n", outputPath)
	return nil
}

func seedEmailNotificationForChaos(ctx context.Context, recipientEmail string) (uuid.UUID, error) {
	recipientEmail = strings.ToLower(strings.TrimSpace(recipientEmail))
	var userID uuid.UUID
	if err := db.Pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, recipientEmail).Scan(&userID); err != nil {
		return uuid.Nil, err
	}

	// Notification must be pending and scheduled_for <= now to be picked up by the worker.
	var insertedID uuid.UUID
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO notifications (user_id, title, body, channel, status, scheduled_for, provider, attempts, last_error)
		VALUES ($1, $2, $3, 'email', 'pending', NOW(), 'sendgrid', 0, '')
		RETURNING id
	`, userID, "Chaos evidence notification", "provider-outage test").Scan(&insertedID)
	if err != nil {
		return uuid.Nil, err
	}
	return insertedID, nil
}

func chaosAssertNotificationDeadLetter(ctx context.Context, notificationID uuid.UUID) (deadLetters int, deliveryAttempts int, finalErr string, err error) {
	err = db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM notification_delivery_attempts WHERE notification_id = $1`, notificationID).
		Scan(&deliveryAttempts)
	if err != nil {
		return 0, 0, "", err
	}
	err = db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM notification_dead_letters WHERE notification_id = $1`, notificationID).
		Scan(&deadLetters)
	if err != nil {
		return 0, deliveryAttempts, "", err
	}
	err = db.Pool.QueryRow(ctx, `SELECT final_error FROM notification_dead_letters WHERE notification_id = $1`, notificationID).
		Scan(&finalErr)
	if err != nil {
		return deadLetters, deliveryAttempts, "", err
	}
	return deadLetters, deliveryAttempts, finalErr, nil
}

func buildSendGridBouncePayload(eventID string, notifID uuid.UUID, recipientEmail string) (payloadBytes []byte, payloadBody string, err error) {
	event := map[string]any{
		"event":         "bounce",
		"timestamp":     time.Now().UTC().Unix(),
		"email":         recipientEmail,
		"reason":        "Mailbox full",
		"sg_message_id": eventID,
		"unique_args": map[string]string{
			"notification_id": notifID.String(),
		},
	}
	payload := []any{event}
	payloadBytes, err = json.Marshal(payload)
	if err != nil {
		return nil, "", err
	}
	payloadBody = string(payloadBytes)
	return payloadBytes, payloadBody, nil
}

func computeSendGridSignature(secret, ts, body string) string {
	mac := hmac.New(sha256.New, []byte(strings.TrimSpace(secret)))
	_, _ = mac.Write([]byte(strings.TrimSpace(ts) + "." + body))
	return hex.EncodeToString(mac.Sum(nil))
}

func chaosAssertSendGridReplay(ctx context.Context, notifID uuid.UUID, eventID string) (eventsCount int, finalErr string, attemptCount int, notifStatus string, err error) {
	err = db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM provider_callback_events
		WHERE provider = 'sendgrid' AND event_id = $1
	`, eventID).Scan(&eventsCount)
	if err != nil {
		return 0, "", 0, "", err
	}

	err = db.Pool.QueryRow(ctx, `
		SELECT final_error, attempt_count
		FROM notification_dead_letters
		WHERE notification_id = $1
	`, notifID).Scan(&finalErr, &attemptCount)
	if err != nil {
		return eventsCount, "", 0, "", err
	}

	err = db.Pool.QueryRow(ctx, `
		SELECT status::text FROM notifications WHERE id = $1
	`, notifID).Scan(&notifStatus)
	if err != nil {
		return eventsCount, finalErr, attemptCount, "", err
	}

	return eventsCount, finalErr, attemptCount, notifStatus, nil
}

func percentile95(vals []int) int {
	if len(vals) == 0 {
		return 0
	}
	cp := make([]int, 0, len(vals))
	cp = append(cp, vals...)
	sort.Ints(cp)
	idx := int(math.Ceil(0.95*float64(len(cp)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(cp) {
		idx = len(cp) - 1
	}
	return cp[idx]
}

func averageInt(vals []int) int {
	if len(vals) == 0 {
		return 0
	}
	sum := 0
	for _, v := range vals {
		sum += v
	}
	return int(math.Round(float64(sum) / float64(len(vals))))
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func writeJSONFile(path string, v any) error {
	dir := filepath.Dir(strings.TrimSpace(path))
	if dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func truncateString(s string, max int) string {
	ss := strings.TrimSpace(s)
	if len(ss) <= max {
		return ss
	}
	return ss[:max] + "..."
}

# Clinical Grade Gap-Closure Stories

This backlog translates remaining clinical-grade gaps into testable user stories.

## CGS-01: Notification Worker Schema Reliability

- **As** a platform operator
- **I want** notification worker queries and writes to run against guaranteed DB schema
- **So that** delivery processing does not fail at runtime in production.

### Acceptance Criteria
- Runtime logs contain no `column n.provider does not exist` worker errors.
- New and existing databases both include required `notifications` columns.
- Delivery telemetry persists to `notification_delivery_attempts`.
- Exhausted retries persist to `notification_dead_letters`.
- `go test ./...` and `npm run cypress:e2e` pass after migration.

## CGS-02: Production-Grade Evidence Expansion

Owner: Rohith (BE)

- **As** a release manager
- **I want** sustained performance and chaos evidence from production-like runs
- **So that** release gates are based on operationally meaningful proof.

### Acceptance Criteria
- Soak test window >= 30 minutes captured in artifact.
- Chaos scenarios include provider outage and delayed callback replay.
- Evidence artifacts include run parameters, observed p95, error-rate, and pass/fail.
- Gate summary links each result to timestamped evidence.

## CGS-03: Usability Validation for Clinical Workflows

Owner: Abhinav (FE)

- **As** a clinical product owner
- **I want** repeatable usability benchmark scenarios with seeded data
- **So that** we can verify that critical workflows are usable under realistic conditions.

### Acceptance Criteria
- Seeded scenario dataset covers patient, doctor, and admin workflows end-to-end.
- Includes AI summary-ready documents and critical-result examples.
- Includes notification, escalation, guardrail, FHIR, and HL7 scenarios.
- Website smoke + E2E runs complete using seeded dataset without manual setup.


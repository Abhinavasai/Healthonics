# Sprint 4 Narrated Video Script (4 Members)

Target duration: 18-24 minutes  
Format: narrated walkthrough with live app + terminal + test evidence + API explorer  
Audience: evaluator or stakeholder unfamiliar with project history

---

## Segment 0 - Opening (00:00-01:00) [All / Host handoff]

Script:

"Welcome to the Sprint 4 presentation for Healthonyx. In this demo, we will show what was built, what was fixed, and how the full-stack platform performs end-to-end across patient, doctor, and admin workflows. We will also show unit test and Cypress evidence, and then close with a complete project pitch as if this is your first time seeing Healthonyx."

---

## Segment 1 - Backend and platform architecture (01:00-06:00) [Kaushik]

Focus:

- Explain role-based architecture:
  - Patient, doctor, admin
  - JWT and RBAC enforcement
- Explain sprint progression:
  - Sprint 1 foundation
  - Sprint 2 geospatial baseline
  - Sprint 3 feature expansion
  - Sprint 4 stabilization and reliability
- Show backend runtime:
  - backend server logs on `:8080`
  - core route groups from `backend/main.go`
- Highlight Sprint 4 reliability fixes:
  - OpenAPI route sync CI fix
  - past appointment prevention logic

Suggested lines:

"The key Sprint 4 backend achievement was to close merge-blocking quality gaps. We made the OpenAPI contract fully consistent with route wiring and added strict future-time validation for appointment create and slot booking so that stale appointments cannot be created."

---

## Segment 2 - Doctor and admin governance workflows (06:00-11:00) [Rohith]

Focus:

- Doctor workflow demo:
  - availability management
  - appointment queue
  - patient documents + download
- Admin workflow demo:
  - lifecycle settings page
  - reason + CONFIRM guardrails
  - AI observability panel
- Sprint 4 issue fix demo:
  - show that short but non-empty reason now works
  - explain backend/UX guardrail alignment

Suggested lines:

"In Sprint 4 we removed the mismatch where the UI accepted a reason but backend rejected it due to a stricter hidden length check. Now both sides consistently enforce non-empty reason plus explicit CONFIRM for sensitive admin actions."

---

## Segment 3 - Patient UX, assistant, notifications, and find-care polish (11:00-16:00) [Abhinav]

Focus:

- Patient workflow demo:
  - find-care location search
  - location autofill/typeahead
  - use-my-location behavior
  - Google map visualization update
  - appointments form preventing past date/time
- Assistant UX:
  - improved floating button design and interaction
  - open assistant and send example prompt
- Notifications:
  - show inbox and preference flow
  - popup behavior and interactive routing mention

Suggested lines:

"Sprint 4 focused heavily on usability in addition to correctness. We improved the assistant entry point, added location autofill, added a clear map visualization, and enforced safer booking behavior directly in the patient form and backend."

---

## Segment 4 - Test evidence and project pitch close (16:00-24:00) [Karthik]

### Part A: unit + e2e evidence

Show terminal outputs for:

- Backend:
  - `cd backend`
  - `go test ./...`
- Frontend:
  - `cd frontend`
  - `npm run test:ci`
- Cypress:
  - `npm run cypress:e2e` (or selected spec run if time constrained)

Call out:

- Sprint 3 test suites are still included and running in current combined suite
- Sprint 4 added/updated tests around appointment safety and find-care Google flow

### Part B: backend API documentation evidence

Show:

- `docs/openapi.yaml`
- `docs/Healthonyx-Sprint4.postman_collection.json`

Explain:

- endpoint coverage by domain (auth, appointments, docs, notifications, admin, messaging, AI, interoperability)
- role protection and request examples

### Part C: final project pitch (for first-time audience)

Suggested close narrative:

"Healthonyx is a full-stack care coordination platform that unifies patient, doctor, and admin operations in one secure workflow. Patients can discover care, manage appointments, receive prescriptions, upload/view records, and message clinicians. Doctors can manage availability, review patient records, and operate appointment queues. Admins can control lifecycle and safety policies, monitor observability, and manage audit/knowledge workflows. The system is API-first with documented contracts, test-backed reliability, and an interoperability path via FHIR and HL7 boundaries."

"Sprint 4 specifically pushed the platform from feature-complete toward deployment-ready by eliminating CI blockers, closing critical workflow bugs, and polishing high-visibility UX surfaces."

---

## Backup Q&A responses (optional)

- Q: "How do you ensure outdated appointments are not booked?"
  - A: "Validation is enforced in both frontend and backend; backend acts as source-of-truth."
- Q: "How do you prove API integrity after many route additions?"
  - A: "OpenAPI sync checks run in CI and fail if route/method drift is detected."
- Q: "Is AI required for core workflows?"
  - A: "No. Core patient/doctor/admin workflows remain operational; AI is additive and configurable."

---

## Recording checklist

- [ ] Backend and frontend are running before recording
- [ ] Demo accounts ready (patient/doctor/admin)
- [ ] Terminal prepared with test commands
- [ ] Postman collection imported before API segment
- [ ] Each member has practiced their time slice
- [ ] Final cut includes all four narrators


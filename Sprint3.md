# Sprint 3 Detailed Report (Healthonyx)

This document records:

1. Detailed Sprint 3 work completed
2. Frontend unit tests inventory
3. Backend unit tests inventory
4. Updated backend API documentation (current route surface)
5. Sprint 4 + Post-Sprint 4 future improvements matrix (as roadmap items)

---

## 1) Sprint 3 Work Completed (Detailed)

## 1.1 Scope evolution from Sprint 2 baseline

Sprint 2 baseline (API + UX) was mostly:

- `POST /api/login`
- `GET /api/geocode`
- `GET /api/hospitals/near`
- `GET /api/doctors/search`
- patient-first find-care UI

By the end of Sprint 3 integration and follow-up merges, the codebase evolved into a role-based clinic platform foundation with patient, doctor, and admin experiences.

## 1.2 Backend work completed in/through Sprint 3

- **Auth + RBAC foundation**
  - JWT auth middleware and role enforcement
  - `/api/register`, `/api/login`, `/api/me`
  - role-restricted patient/doctor/admin endpoints

- **Find-care and booking foundation**
  - geocoding endpoint
  - nearby hospitals and doctor search
  - doctor slots create/delete/list and patient slot booking

- **Appointments**
  - create/list/detail APIs
  - doctor status updates
  - patient cancel pathway
  - appointment activity timeline APIs
  - appointment comments APIs

- **Dashboard APIs**
  - patient summary endpoint
  - doctor summary endpoint

- **Documents and records**
  - patient documents list/upload/download
  - doctor document list/detail/summarize route path
  - patient files list/upload/download

- **Prescriptions**
  - prescriptions list by patient
  - doctor create prescription
  - doctor/admin revoke prescription

- **Messaging**
  - unread total
  - list threads
  - create thread
  - list messages in thread
  - send message
  - mark thread read

- **Notifications**
  - in-app notifications list endpoint

- **Admin**
  - admin protected endpoint
  - audit log list endpoint
  - knowledge docs list/create/get endpoints

## 1.3 Frontend work completed in/through Sprint 3

- **Role-based app shell + route guards**
  - guarded routes for patient/doctor/admin
  - role-aware navigation and page access

- **Patient UI**
  - dashboard
  - find-care
  - appointments list + detail
  - patient documents
  - messages
  - my files
  - prescriptions
  - notifications inbox

- **Doctor UI**
  - dashboard
  - availability
  - appointments list + detail
  - doctor documents list + detail
  - messages
  - notifications inbox

- **Admin UI**
  - admin shell
  - audit view
  - knowledge management view

## 1.4 Integration and quality achievements

- End-to-end branch integration strategy documented (`docs/sprint3-merge-order.md`)
- automated backend tests (`go test ./...`)
- frontend unit tests (`npm run test:ci`)
- full frontend build validation (`npm run build`)
- Cypress E2E suite execution (`npm run cypress:e2e`)
- Postman + Newman collection setup for API verification

---

## 2) Frontend Unit Tests (Current Inventory)

Located under `frontend/src/.../*.spec.ts`:

1. `frontend/src/app/components/doctor-dashboard/doctor-dashboard.component.spec.ts`
2. `frontend/src/app/services/documents.service.spec.ts`
3. `frontend/src/app/components/admin-audit/admin-audit.component.spec.ts`
4. `frontend/src/app/components/patient-find-care/patient-find-care.component.spec.ts`
5. `frontend/src/app/components/patient-prescriptions/patient-prescriptions.component.spec.ts`
6. `frontend/src/app/components/admin-knowledge/admin-knowledge.component.spec.ts`
7. `frontend/src/app/components/patient-dashboard/patient-dashboard.component.spec.ts`
8. `frontend/src/app/services/geo-booking.service.spec.ts`
9. `frontend/src/app/services/doctor-documents.service.spec.ts`
10. `frontend/src/app/services/messaging.service.spec.ts`
11. `frontend/src/app/components/notifications-inbox/notifications-inbox.component.spec.ts`
12. `frontend/src/app/components/patient-files/patient-files.component.spec.ts`
13. `frontend/src/app/components/doctor-appointments/doctor-appointments.component.spec.ts`

Run command:

```bash
cd frontend
npm run test:ci
```

---

## 3) Backend Unit Tests (Current Inventory)

Located under `backend/handlers/*_test.go`:

1. `backend/handlers/patient_files_test.go`
2. `backend/handlers/appointment_comments_test.go`
3. `backend/handlers/dashboard_test.go`
4. `backend/handlers/documents_test.go`
5. `backend/handlers/geo_booking_test.go`
6. `backend/handlers/admin_audit_test.go`
7. `backend/handlers/doctor_documents_test.go`
8. `backend/handlers/bootstrap_test.go`
9. `backend/handlers/notifications_test.go`
10. `backend/handlers/prescriptions_test.go`
11. `backend/handlers/knowledge_admin_test.go`
12. `backend/handlers/geocode_handler_test.go`
13. `backend/handlers/messaging_test.go`
14. `backend/handlers/appointments_lifecycle_test.go`

Run commands:

```bash
cd backend
go test ./...
go build ./...
```

---

## 4) Updated Backend API Documentation (Current Route Surface)

Source of truth: `backend/main.go` (current `dev` wiring).

## 4.1 Health

- `GET /health`
- `HEAD /health`

## 4.2 Auth and bootstrap

- `POST /api/register`
- `POST /api/login`
- `GET /api/me`
- `GET /api/bootstrap`

## 4.3 Notifications

- `GET /api/notifications`

## 4.4 Dashboards

- `GET /api/patient/dashboard/summary` (patient)
- `GET /api/doctor/dashboard/summary` (doctor)

## 4.5 Doctor documents

- `GET /api/doctor/documents` (doctor)
- `GET /api/doctor/documents/:id` (doctor)
- `POST /api/doctor/documents/:id/summarize` (doctor)

## 4.6 Admin

- `GET /api/admin`
- `GET /api/admin/audit-log`
- `GET /api/admin/knowledge-docs`
- `POST /api/admin/knowledge-docs`
- `GET /api/admin/knowledge-docs/:id`

## 4.7 Role validation endpoints

- `GET /api/doctor`
- `GET /api/patient`

## 4.8 Find-care and geo booking

- `GET /api/geocode` (patient)
- `GET /api/hospitals/near` (patient)
- `GET /api/doctors/search` (patient)
- `GET /api/doctors/:id/slots` (patient)
- `POST /api/doctor/slots` (doctor)
- `DELETE /api/doctor/slots/:id` (doctor)
- `POST /api/appointments/book-slot` (patient)

## 4.9 Appointments

- `POST /api/appointments` (patient)
- `POST /api/patient/appointments/:id/cancel` (patient)
- `GET /api/appointments/patient` (patient)
- `GET /api/appointments/doctor` (doctor)
- `GET /api/appointments/:id/activity`
- `GET /api/appointments/:id/comments`
- `POST /api/appointments/:id/comments`
- `GET /api/appointments/:id`
- `GET /api/doctors` (patient doctor list)
- `PATCH /api/appointments/:id/status` (doctor)

## 4.10 Patient documents and files

- `GET /api/documents` (patient)
- `POST /api/documents` (patient)
- `GET /api/documents/:id/download` (patient)
- `GET /api/patients/:patientId/files`
- `POST /api/patients/:patientId/files`
- `GET /api/files/:id`

## 4.11 Prescriptions

- `GET /api/patients/:patientId/prescriptions`
- `POST /api/patients/:patientId/prescriptions` (doctor)
- `PATCH /api/prescriptions/:id/revoke` (doctor/admin)

## 4.12 Messaging

- `GET /api/messages/unread` (patient/doctor)
- `GET /api/messages/threads` (patient/doctor)
- `POST /api/messages/threads` (patient/doctor)
- `GET /api/messages/threads/:threadId` (patient/doctor)
- `POST /api/messages/threads/:threadId/messages` (patient/doctor)
- `POST /api/messages/threads/:threadId/read` (patient/doctor)

---

## 5) Sprint 4 + Post-Sprint 4 Matrix (Future Improvements Roadmap)

> The section below is intentionally included as future roadmap detail (as requested).  
> It is not a claim that all items are currently implemented.

Below is a separate, detailed matrix for what still needs to be developed against the PDF plus the post-AI roadmap.

## 5.1 Documents and Report Uploads

Status now (roadmap framing): Missing as a real module.

| Sub-area | Current status | What still needs to be built |
|---|---|---|
| DB schema | Missing | `attachments` / `patient_files` table with id, patient_id, uploaded_by, description, file_path, mime_type, size, status, timestamps, optional record_id |
| File storage | Missing | Local uploads dir or object storage abstraction, secure file naming, path persistence |
| Upload API | Missing | `POST /api/patients/{patientId}/files` multipart upload |
| List API | Missing | `GET /api/patients/{patientId}/files` |
| Download/view API | Missing | `GET /api/files/{fileId}` with auth checks |
| Visibility rules | Missing | Patient sees own docs, treating doctors see permitted docs, admin access policy |
| Doctor upload flow | Missing | Doctor-side patient file upload UI and backend |
| Patient upload flow | Missing | Patient upload UI with description |
| Preview UX | Missing | PDF/image preview or secure download |
| Review workflow | Missing | `pending_review` / `available` / approval rules |
| Upload history | Missing | Uploader, date, description, audit log |
| Security | Missing | File type/size validation, optional ClamAV, secure storage paths |

## 5.2 Prescription Module

Status now (roadmap framing): Missing.

| Sub-area | Current status | What still needs to be built |
|---|---|---|
| DB schema | Missing | `prescriptions` table with patient, doctor, med, dosage, frequency, duration, instructions, status |
| History/audit | Missing | Change tracking or superseded records |
| Create/update/revoke API | Missing | Doctor-only endpoints |
| Read APIs | Missing | Patient and doctor list/detail endpoints |
| Doctor UI | Missing | Create/update/revoke forms in patient context |
| Patient UI | Missing | “My Prescriptions” page with active/past states |
| PDF generation | Missing | Prescription print/download document |
| Reminder linkage | Missing | Scheduling metadata per prescription |
| Per-prescription reminder toggle | Missing | Boolean or preferences override per prescription |
| Notification on change | Missing | Send/log when prescription updated/revoked |

## 5.3 Notifications and Delivery Controls

Status now: Partial (preferences exist; delivery system does not).

| Sub-area | Current status | What still needs to be built |
|---|---|---|
| User prefs DB | Basic global prefs exist | Expand to categories and per-item overrides |
| Preferences UI | Basic email/SMS appointment prefs | Add medication reminders, new lab/doc alerts, announcements, per-prescription/report toggles |
| Delivery scheduler | Missing | Cron/worker for due notifications |
| Notifications table/log | Missing | `notifications` table with scheduled/sent/failed status |
| Appointment reminders | Missing | Schedule on appointment create/update |
| Medication reminders | Missing | Schedule on prescription create/update |
| Email integration | Missing | SendGrid/Mailgun/SMTP integration |
| SMS integration | Missing | Twilio integration |
| Admin monitoring | Missing | Notification log viewer, failure filtering, queue counts |
| Global delivery settings | Missing | Reminder lead time, channel defaults, emergency disable |
| Compliance flow | Missing | Opt-in/out policy, footer copy, retry/failure handling |

## 5.4 Appointment Scheduling System

Status now: Strong partial.

| Sub-area | Current status | What still needs to be built |
|---|---|---|
| Doctor availability | Implemented | Calendar-style UX if desired |
| Patient booking | Implemented | Alternate-doctor suggestion / richer slot UX |
| Conflict prevention | Partly implemented via slots | Explicit patient-doctor overlap rules and better error messaging |
| Status lifecycle | Implemented strongly | Confirm exact parity with PDF naming and audit breadth |
| Reschedule/cancel | Implemented | Cutoff rules, cancellation notifications |
| Confirmation notifications | Missing delivery infra | Send/log patient and doctor confirmations |
| Appointment reminders | Missing delivery infra | 24h reminders, configurable lead time |
| Calendar UI | Partial | Full weekly/day calendar with richer visualization |
| Doctor-side direct scheduling | Limited | Explicit doctor “add appointment for patient” workflow |
| No-show analytics | Basic status exists | Dashboard metrics and admin reporting |

## 5.5 Dashboards

Status now: Partial.

| Sub-area | Current status | What still needs to be built |
|---|---|---|
| Patient counts/upcoming | Implemented | — |
| Patient next appointment | Implemented | — |
| Patient health metrics charts | Missing | BP/sugar/weight trends and chart widgets |
| Active prescriptions widget | Missing | Depends on prescription module |
| New reports/documents alert | Missing | Depends on documents module |
| Health insights/rule-based tips | Missing | Reminder nudges, annual exam alerts, care gaps |
| Doctor today/pending/unread | Implemented | — |
| Doctor patient census | Missing | Count/overview of assigned patients |
| Doctor alerts | Missing | Critical labs, follow-up needed, docs needing attention |
| Workload charts | Missing | Appointments/week, no-show rate, patient mix |
| Doctor AI summaries | Missing | Appointment contextual one-liners from records/docs |
| Admin metrics | Basic counts and DB health | Appointments, prescriptions, signups, engagement, notification KPIs |

## 5.6 Secure Messaging and Comments

Status now: Messaging implemented; comments missing; realtime missing.

| Sub-area | Current status | What still needs to be built |
|---|---|---|
| Direct patient-doctor messaging | Implemented | — |
| Thread list/inbox | Implemented | — |
| Read/unread | Implemented | — |
| WebSocket realtime | Missing | Live updates with fallback polling |
| Email fallback for unread | Missing | Notification hook on unread threshold |
| Comments on records/docs | Missing | `comments` table or contextual comments model |
| Comment visibility | Missing | Internal vs patient-visible |
| Comment notifications | Missing | Notify on shared comment |
| Knowledge-doc comments | Missing | Internal collaboration thread per KB doc |
| Stronger provider-patient scoping | Partial | “Only own patients” restrictions beyond simple role pairing |

## 5.7 Knowledge Base Health Monitor

Status now: Missing.

| Sub-area | Current status | What still needs to be built |
|---|---|---|
| Knowledge docs table | Missing | `knowledge_docs` |
| Versioning table | Missing | `knowledge_doc_versions` |
| Doc editor APIs | Missing | CRUD + version bump |
| Angular KB UI | Missing | List, editor, version history, analytics |
| Freshness threshold config | Missing | Admin setting for stale threshold |
| Staleness jobs | Missing | Periodic review scanner |
| Health score | Missing | Freshness + usage + conflict composition |
| Similarity detection | Missing | Embeddings + vector search |
| Conflict alert table | Missing | `document_alerts` / conflicts |
| AI contradiction analysis | Missing | LLM compare pass for candidate conflicts |
| Admin analytics dashboard | Missing | Stale docs, conflict count, score distribution |
| Public/private visibility | Missing | Optional patient-visible educational docs |

## 5.8 Admin Panel and System Management

Status now: Partial.

| Sub-area | Current status | What still needs to be built |
|---|---|---|
| User list + active toggle | Implemented | — |
| Create/edit users | Missing | Admin create user forms and APIs |
| Doctor-specific profile fields | Partial at DB level | Admin UI/backend for specialization/license/department fields |
| Password reset/recovery | Missing | Admin-triggered reset flow |
| System dashboard | Basic counts/DB ok | More KPIs, trends, signups, no-shows, prescriptions |
| Audit logs | Missing as general system feature | Central audit log table + viewer |
| Global settings | Missing | Reminder times, KB thresholds, self-service toggles, provider credentials |
| Notification monitoring | Missing | Log viewer and failed sends |
| KB monitoring integration | Missing | Depends on KB module |
| AI settings | Missing | Enable flags, model names, limits, reindex actions |
| Deployment/admin ops | Missing in-product | System version/build info, service health surface |

## 5.9 Post-Sprint 4 AI Platform Matrix

### Phase A: AI Foundation (Status now: Missing)

| Area | What to build |
|---|---|
| Config | `AI_ENABLED`, `RAG_ENABLED`, `AI_MODEL`, `EMBEDDING_MODEL`, `VECTOR_BACKEND`, timeouts |
| AI service | Separate service/module with health/version endpoints |
| Internal auth | Shared secret / mTLS from Go API to AI service |
| Compose/deploy | Docker compose for API + Ollama + Qdrant + AI worker |
| Audit fields | `ai_model`, `ai_prompt_version`, `source_chunk_ids`, timestamp |
| Contract tests | Go <-> AI service tests |
| Fallback behavior | App remains fully usable with AI off |

### Phase B: Sprint 4 Light AI (Status now: Missing)

| Track | What to build |
|---|---|
| Find-care rerank | Embedding-based rerank for free-text symptom search |
| Dashboard digest | Structured-data summaries, no diagnosis |
| Prescription summary | Plain-language patient-facing explanation |
| Admin AI controls | Toggle AI, choose model, set limits |
| Notification paraphrase | Optional LLM rewording for reminders |

### Phase C: KB + RAG (Status now: Missing)

| Area | What to build |
|---|---|
| Chunking pipeline | Split KB docs into chunks |
| Embeddings pipeline | Generate/store vectors |
| Vector storage | Qdrant/FAISS/pgvector |
| Hybrid retrieval | FTS + vector |
| RAG endpoint | Query -> answer + citations |
| Conflict workflow | Similarity candidates + LLM contradiction notes |
| Reindex jobs | Async/manual reindex |

### Phase D: Document AI Summaries (Status now: Missing)

| Area | What to build |
|---|---|
| Text extraction | PDF text + OCR for images |
| Summary jobs | Async summarization queue |
| Summary status | `pending` / `ready` / `failed` |
| UI | “Summarizing…” then show summary and sources |
| Safety | Doctor review before patient-facing summaries if required |

### Phase E: Hardening (Status now: Missing for AI-specific concerns)

| Area | What to build |
|---|---|
| Rate limiting | Per-user AI usage caps |
| Caching | Embedding cache, RAG answer cache |
| Eval harness | Fixed fixtures, hallucination/citation checks |
| Observability | p95 latency, queue depth, failures |
| Prompt safety | Injection tests, truncation and input limits |
| Security posture | Sandboxed worker, restricted outbound access if needed |

---

## 6) Practical summary

Compared to Sprint 2, the project now includes:

- full auth + role-based shell
- appointments + activity + doctor availability
- messaging
- dashboards
- admin basics
- notification preferences/inbox
- richer find-care funnel
- tests + Postman + Cypress support

Largest roadmap blocks still pending are:

- full AI platform + RAG + KB health monitor
- robust delivery infrastructure (email/SMS workers/retries)
- enterprise-level admin and analytics expansion


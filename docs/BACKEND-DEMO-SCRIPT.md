# Backend Demo Script (Simple English)

Use this as a speaking script during backend demo, viva, or project review.

---

## 1) Opening statement (30-60 seconds)

"In Sprint 2, our backend was mainly a discovery flow: login, geocode, nearby hospitals, and doctor search.  
From Sprint 3 onward, we expanded it into a role-based clinic backend with patient, doctor, and admin APIs.  
Today I will show the backend progress from Sprint 2 to current state, then run Postman to verify feature coverage."

---

## 2) Sprint 2 vs current backend progress

### Sprint 2 baseline

- `POST /api/login`
- `GET /api/geocode`
- `GET /api/hospitals/near`
- `GET /api/doctors/search`

### What we added after Sprint 2

- Auth + role context: register/login/me, role-protected routes.
- Dashboards:
  - `GET /api/patient/dashboard/summary`
  - `GET /api/doctor/dashboard/summary`
- Appointments:
  - create/list/detail/status update
  - activity timeline
  - comments on appointments
  - patient cancel flow
- Doctor availability + slot booking:
  - create/delete slot
  - list open slots
  - book slot
- Documents and files:
  - patient documents
  - doctor document list/detail/summarize endpoint path
  - patient file list/upload/download
- Prescriptions:
  - list/create/revoke endpoints
- Messaging:
  - unread count, thread list, create thread, send, mark read
- Notifications:
  - inbox endpoint
- Admin:
  - admin auth route
  - audit-log endpoint
  - knowledge docs list/create/detail

---

## 3) Backend architecture explanation (simple)

"Our backend is a Go API using Gin.  
`main.go` wires all route groups under `/api`, and each route goes to a focused handler module.  
Authentication is JWT-based.  
Authorization is role-based at route level and business-level checks are done inside handlers for ownership/access.  
Database migration runs at startup so schema changes are applied automatically."

Use these examples while speaking:

- Role guard at route level: patient/doctor/admin access split.
- Business guard in handlers:
  - doctor can update only their appointments
  - patient can access only own records
  - messaging only allowed patient-doctor pair

---

## 4) What is complete vs partial (backend view)

### Strongly implemented

- Core auth and role protection.
- Find-care APIs (geocode/hospitals/doctors).
- Appointment lifecycle base + comments/activity.
- Messaging basic workflow.
- Dashboard summary endpoints.
- Basic admin audit/knowledge endpoints.
- Prescriptions and patient files APIs.

### Partial / still planned

- Full notification delivery engine (email/SMS workers/queue).
- AI platform pieces (Ollama + embeddings + vector DB + RAG pipeline).
- Full KB health monitor and contradiction workflow.
- Rich admin operations (global settings and advanced analytics).

---

## 5) Backend demo run commands

From repo root:

```powershell
docker compose up -d
cd backend
go run .
```

Optional seed (for predictable demo data):

```powershell
go run ./cmd/seed
```

Backend test commands:

```powershell
go test ./...
go build ./...
```

---

## 6) Postman demo script (important)

Say:

"Now I will validate backend APIs using our full Postman collection and local environment."

### Files to import in Postman

- `docs/Healthonyx-API-Full.postman_collection.json`
- `docs/Healthonyx-Local.postman_environment.json`

### Runner order

Run collection folders in this order:

1. `00 - Smoke (Runnable)`
2. `01 - Dashboards`
3. `02 - Find Care + Slots`
4. `03 - Appointments + Comments`
5. `04 - Documents + Files + Prescriptions`
6. `05 - Messaging + Notifications`
7. `06 - Admin`

### What this Postman setup does automatically

- Logs in patient, doctor, admin.
- Stores tokens into variables.
- Derives key IDs from API responses:
  - `doctorId`
  - `appointmentId`
  - `threadId`
  - `prescriptionId`
  - optional `documentId` if available

### Newman CLI equivalent

```powershell
npx newman run docs/Healthonyx-API-Full.postman_collection.json `
  -e docs/Healthonyx-Local.postman_environment.json `
  --env-var baseUrl=localhost:8080
```

---

## 7) Common demo troubleshooting lines

If 401 appears:

"This means token was not set or Smoke folder was skipped. I rerun folder 00 first."

If 404 appears:

"This usually means wrong backend instance/port or missing seed data for that specific record."

If 409 appears (slots):

"This is expected when creating a duplicate slot time."

---

## 8) Closing statement

"Compared to Sprint 2, backend has moved from simple find-care APIs to a multi-role clinic platform foundation with appointments, messaging, documents/files, prescriptions, notifications, and admin modules.  
The next major backend milestones are AI/RAG integration, delivery infrastructure, and enterprise-grade admin controls."


# Sprint 3 — Submission Document (detailed)

This file supports **course submission**: paste the **GitHub** link in the LMS, use **comments** for **multiple video URLs** if the form only allows one primary link, and attach or reference this document where your instructor allows.

---

## 1. Repository and branch

| Item | Value |
|------|--------|
| **GitHub repository** | [https://github.com/Abhinavasai/Healthonyx](https://github.com/Abhinavasai/Healthonyx) |
| **Sprint 3 integration branch** | `s3-feature-updates-healthonyx` |
| **Related planning docs** | `docs/sprint3-assigned-features.md`, `docs/sprint3-merge-order.md`, `docs/BRANCHES.md` |

Graders may review **per-commit authorship** on this branch (and feature branches merged into it). Each teammate should have **distinct commits** with meaningful messages.

---

## 2. What to submit on the LMS

| Field / area | What to enter |
|----------------|----------------|
| **Primary URL** | The GitHub repo link above (or the branch URL if your rubric requires it). |
| **Comments / additional links** | Links to **narrated videos** (Part 1, Part 2, unlisted YouTube, Drive with sharing, etc.). |
| **Attachments** | If allowed, export this file as PDF or link to **Sprint3.md** on the `main`/`dev`/integration branch **tagged at submission time**. |

---

## 3. Video presentation — rubric mapping

Your instructor asked for a **narrated** recording that **splits narration across team members** and covers specific artifacts. Use this as a **shot list** so nothing is missed.

### 3.1 Must cover (checklist)

| # | Requirement | Practical demonstration |
|---|-------------|-------------------------|
| 1 | **Each member narrates a portion** | Record in segments. Example split: (A) intro + backlog, (B) backend + Postman walkthrough, (C) Angular UI demos, (D) unit tests + docs + wrap-up. |
| 2 | **Demonstrate new Sprint 3 functionality** | Live UI walkthrough: dashboards, appointment detail + activity, patient/doctor appointment actions (cancel, reschedule, approve/reject, complete/no-show as implemented), find-care flow, messaging, availability/slots, admin user list + active toggle, notification preferences. |
| 3 | **Show results of all unit tests (incl. Sprint 2)** | Clear terminal: `go test ./...` from `backend/`, then `npx ng test --no-watch --browsers=ChromeHeadlessNoSandbox` from `frontend/`. Explain that **geo + messaging Angular specs** are the Sprint 2 carryover suites still in the repo. |
| 4 | **This document (`Sprint3.md`)** | Scroll through sections **1–7** or share screen in VS Code / GitHub raw view for ~30–60 seconds. |
| 5 | **Detail Sprint 3 work** | Walk through **Section 5** (feature breakdown) verbally with short demos tied to bullets. |
| 6 | **List frontend unit tests** | Show **Section 6** table or expand Karma output so individual **`it()`** names are visible. |
| 7 | **List backend unit tests** | Show **Section 7** table or `go test -v ./...` (may be long—scroll through handlers + models). |
| 8 | **Updated backend API documentation** | Open Postman (imported collections) and/or **Section 8** endpoint matrix; run **one** Newman example if you want extra credit style polish. |

### 3.2 Suggested time structure (adjust to your max length)

| Segment | Owner (fill names) | Content |
|---------|-------------------|---------|
| 0:00–1:30 | Member 1 | Problem recap, repo link, branch name, demo environment (DB + seed). |
| 1:30–6:00 | Member 2 | Backend: REST highlights + **Postman** folder walk. |
| 6:00–11:00 | Member 3 | Frontend: patient + doctor + admin paths. |
| 11:00–14:00 | Member 4 | **Unit tests** (both stacks) + quick Cypress/Newman mention if time. |
| 14:00–15:00 | Any | Scroll `Sprint3.md`, thanks, commit/collab note. |

### 3.3 Video URL placeholders

Replace with your uploaded links before submitting:

- **Recording (full or Part 1):** `https://`
- **Part 2 (optional):** `https://`
- **Unlisted / Drive notes:** *(permissions: “anyone with link” if graders are external)*

---

## 4. Local environment and “how we ran it”

Use this so your **recording** matches what graders can reproduce.

### 4.1 Prerequisites

- **PostgreSQL** with an empty or seeded database.
- **Go** (module in `backend/go.mod`).
- **Node.js** + npm (see `frontend/package.json`).

### 4.2 Backend environment (`backend/.env` or system env)

| Variable | Purpose |
|----------|---------|
| `DATABASE_URL` | Postgres connection string (required). |
| `JWT_SECRET` | Signing key for JWTs (required). |
| `PORT` | Default **8080** if unset. |
| `CORS_ORIGINS` | Optional; comma-separated (default allows `http://localhost:4200`). |
| `NOMINATIM_BASE_URL` | Optional override for geocoding. |
| `GEOCODE_USER_AGENT` | Recommended for Nominatim good citizenship. |

### 4.3 Seed data (typical demo)

From repo root (adjust paths):

```powershell
cd backend\cmd\seed
go run .
```

Demo users are printed by the seed (commonly `patient@healthonyx.demo`, `doctor@healthonyx.demo`, `admin@healthonyx.demo` with passwords in seed output — **use your terminal output as source of truth**).

Optional extra fixtures:

```powershell
cd backend\cmd\seed_testdata
go run .
```

### 4.4 Run the API

```powershell
cd backend
go run .
```

Expect: `Listening on :8080` (or your `PORT`).

### 4.5 Run the SPA (with `/api` proxy)

```powershell
cd frontend
npm install
ng serve --host 127.0.0.1 --port 4200
```

Open **http://localhost:4200**. Full manual checklist: **`docs/MANUAL-TESTING.md`**.

---

## 5. Work completed in Sprint 3 (detailed narrative)

Sprint 3 builds on **Sprint 1–2** (auth, role shell, geo search, messaging foundations) and adds **operational** features: visibility into single appointments, **audit trail**, **dashboards**, **admin**, **preferences**, and richer **booking lifecycle** integrated with slots and geo.

Customize the bullets below with **your** ticket/story IDs and owner names if the rubric asks for attribution.

### 5.1 Backend — REST API and behavior

- **Health:** `GET /health` for uptime checks.
- **Auth:** `POST /api/register`, `POST /api/login`, `GET /api/me` (JWT). Inactive users are rejected on authenticated routes when `active` is false in the DB.
- **Bootstrap:** `GET /api/bootstrap` — app config for the shell (e.g. mobile breakpoint).
- **Notification preferences:** `GET` and `PUT /api/me/notification-preferences` — persisted per-user settings.
- **Dashboards (aggregated metrics):**
  - `GET /api/dashboard/patient` (role: patient)
  - `GET /api/dashboard/doctor` (role: doctor)
  - `GET /api/dashboard/admin` (role: admin)
- **Admin users:** `GET /api/admin/users`, `PATCH /api/admin/users/:id` — list users and toggle `active` / other allowed fields per handler.
- **Role probes:** `GET /api/admin`, `/api/doctor`, `/api/patient` — intended to demonstrate **403** when JWT role does not match.
- **Geo & booking (patient):** `GET /api/geocode`, `GET /api/hospitals/near`, `GET /api/hospitals/:id/departments`, `GET /api/doctors/search`, `GET /api/doctors/:id/slots`, `POST /api/appointments/book-slot`.
- **Doctor availability:** `POST /api/doctor/slots`, `DELETE /api/doctor/slots/:id`.
- **Appointments:** `POST /api/appointments`; `GET /api/appointments/patient` / `GET /api/appointments/doctor`; `GET /api/appointments/:id` (**detail** — patient/doctor scoping); `GET /api/appointments/:id/activity` (**activity log**); `PATCH /api/appointments/:id/status` (doctor); patient-driven `PATCH .../cancel` and `PATCH .../request-reschedule`; `GET /api/doctors` (patient directory for manual booking flows).

- **Messaging (patient + doctor only):** `GET/POST` threads, list messages, send, mark read, unread total, compose limits — under `/api/messages/...`.

### 5.2 Data and migrations

- Migrations run **on startup** via `db.Migrate` (see `backend/main.go`). They include tables/columns needed for **appointment activities**, **dashboards**, **preferences**, **admin flags**, etc., as implemented on the integration branch.

### 5.3 Frontend — routes and UX (high level)

Paths are guarded by **Angular route guards** by role (exact child paths may vary; search `app.routes` / `app.config` in `frontend/src` for the canonical list).

- **Patient:** dashboard, **find care** (geocode → hospitals → departments → doctors → slots/booking), **appointments** list and **detail** with **activity timeline**, **messaging**, **settings** (account + notification preferences).
- **Doctor:** dashboard, **appointments** queue and detail with actions (approve/reject, status updates, reschedule handling, completion/no-show as implemented), **availability** / slots, **messaging**, **settings**.
- **Admin:** dashboard with DB health and aggregates, **user administration** table with enable/disable.

### 5.4 Quality, regression, and documentation

- **Postman:** Sprint 3 targeted collection + full regression collection under `docs/` (import and set `baseUrl`).
- **Manual guide:** `docs/MANUAL-TESTING.md`.
- **Test matrices:** `docs/TEST-CASES-Sprint3-Implemented.md`, `docs/TEST-CASES-Frontend-Backend-E2E-Complete.md`.
- **Karma:** `frontend/karma.conf.js` + `angular.json` test target for **ChromeHeadlessNoSandbox** — stabilizes headless runs on CI and Windows.

---

## 6. Frontend unit tests (complete list)

**Run:**

```powershell
cd frontend
npx ng test --no-watch --browsers=ChromeHeadlessNoSandbox
```

**Expected:** **12** passing specs (current repository state).

### 6.1 Spec files and what they validate

| Spec file | Count | Focus |
|-----------|-------|--------|
| `src/app/components/patient-find-care/patient-find-care.component.spec.ts` | 6 | Search UX: empty query error, geocode selection, geolocation error path, hospitals/doctors API wiring. |
| `src/app/services/geo-booking.service.spec.ts` | 2 | HTTP calls to `/api/geocode` and `/api/hospitals/near` with expected query params (Sprint 2 geo stack). |
| `src/app/services/messaging.service.spec.ts` | 4 | Thread list/unread, `refreshUnread`, send message, create thread (Sprint 2 messaging stack). |

### 6.2 Per-test inventory (names as in source)

| # | File | Test name (`it(...)`) |
|---|------|----------------------|
| 1 | `patient-find-care.component.spec.ts` | shows an error if locationQuery is empty on search |
| 2 | `patient-find-care.component.spec.ts` | runs geocodeSearch and updates lat/lng + label for a single hit |
| 3 | `patient-find-care.component.spec.ts` | pickGeocodeSuggestion updates locationQuery and search point |
| 4 | `patient-find-care.component.spec.ts` | sets permission denied error when geolocation fails |
| 5 | `patient-find-care.component.spec.ts` | loadHospitals calls API with current lat/lng/radius and sets hospitals |
| 6 | `patient-find-care.component.spec.ts` | loadDoctors sets doctors from API response |
| 7 | `geo-booking.service.spec.ts` | geocodeSearch should call /api/geocode with q and limit |
| 8 | `geo-booking.service.spec.ts` | hospitalsNear should call /api/hospitals/near with lat/lng/radius_km |
| 9 | `messaging.service.spec.ts` | lists threads and updates unread from thread rows |
| 10 | `messaging.service.spec.ts` | refreshUnread uses /api/messages/unread |
| 11 | `messaging.service.spec.ts` | sendMessage posts to thread messages |
| 12 | `messaging.service.spec.ts` | createThread posts peer and body |

**Note:** Sprint 3 UI for dashboards and appointment detail is covered primarily by **Cypress E2E** and **manual** checks unless you add new `*.spec.ts` files later. The **12 tests** are the current **automated unit** count required for the “show unit tests” portion of the rubric.

---

## 7. Backend unit tests (complete list)

**Run (default — integration tests skip without env):**

```powershell
cd backend
go test ./... -count=1
```

**Packages with tests:** `handlers`, `models`.

### 7.1 Integration tests (optional full run)

These tests **exit early with `t.Skip`** unless you set environment variables (see file headers):

| Test | Enable with |
|------|-------------|
| `TestMe_reloadsFromDatabase` | `RUN_AUTH_ME_INTEGRATION=1` + `DATABASE_URL` |
| `TestAppointmentActivity_flow` | `RUN_APPOINTMENT_ACTIVITY_INTEGRATION=1` + `DATABASE_URL` |

Example:

```powershell
$env:DATABASE_URL="postgres://USER:PASSWORD@localhost:5432/healthonyx?sslmode=disable"
$env:RUN_AUTH_ME_INTEGRATION="1"
$env:RUN_APPOINTMENT_ACTIVITY_INTEGRATION="1"
cd backend
go test ./handlers -v -count=1 -run "TestMe_reloadsFromDatabase|TestAppointmentActivity_flow"
```

### 7.2 Per-file inventory

| File | Functions |
|------|-----------|
| `handlers/auth_me_test.go` | `TestMe_UnauthorizedWithoutClaims` |
| `handlers/auth_me_integration_test.go` | `TestMe_reloadsFromDatabase` |
| `handlers/bootstrap_test.go` | `TestBootstrap_UnauthorizedWithoutClaims`, `TestBootstrap_OKWithClaims` |
| `handlers/geocode_handler_test.go` | `TestGeocodeSearch_MissingQ`, `TestGeocodeSearch_ParsesUpstreamResults` |
| `handlers/geo_booking_test.go` | `TestHaversineKm_ZeroDistance`, `TestParseLatLngRadius_ValidDefaults`, `TestParseLatLngRadius_MissingLat`, `TestParseLatLngRadius_OutOfBounds` |
| `handlers/messaging_test.go` | `TestPreviewText_ShortUnchanged`, `TestPreviewText_TruncatesLong`, `TestValidPatientDoctorPair`, `TestComposeLimits_OK`, `TestComposeLimits_Unauthorized`, `TestUnreadTotal_Unauthorized`, `TestListMessages_InvalidThreadUUID`, `TestSendMessage_InvalidJSON`, `TestCreateThread_InvalidPeerUUID` |
| `handlers/appointments_activity_integration_test.go` | `TestAppointmentActivity_flow` |
| `models/appointment_activity_test.go` | `TestAppointmentActivity_JSONFields` |

**Coverage themes:** JWT/bootstrap contracts, **geocode + geo query parsing**, **messaging** validation and limits, **appointment activity** model and end-to-end DB flow when integration env is set.

---

## 8. Backend API documentation (artifacts and endpoint index)

### 8.1 Machine-readable collections (import into Postman or run with Newman)

| File | Purpose |
|------|---------|
| `docs/Healthonyx-API-Sprint3-Implemented.postman_collection.json` | Sprint 3 focused requests + tests |
| `docs/Healthonyx-API-Complete-Regression.postman_collection.json` | Broader regression |
| `docs/Healthonyx-API-Sprint2.postman_collection.json` | Sprint 2 geo reference |

**Newman (CLI):**

```powershell
cd path\to\Healthonyx
npx newman run docs\Healthonyx-API-Sprint3-Implemented.postman_collection.json
npx newman run docs\Healthonyx-API-Complete-Regression.postman_collection.json
```

Set collection variable **`baseUrl`** = `http://localhost:8080` (or your host). Login requests typically set **`token`** for downstream calls.

### 8.2 Human-readable docs

| File | Contents |
|------|----------|
| `docs/MANUAL-TESTING.md` | Step-by-step local run and manual feature verification |
| `docs/TEST-CASES-Sprint3-Implemented.md` | FE/BE/E2E case IDs |
| `docs/TEST-CASES-Frontend-Backend-E2E-Complete.md` | Expanded regression matrix |
| `docs/geo-specialty-booking.md` | Backend geo/booking notes |
| `docs/geo-specialty-booking-frontend.md` | Frontend integration notes |

### 8.3 Authoritative route index (from `backend/main.go`)

Use this table in the video when you “show API documentation.”

| Method | Path | Auth / role | Notes |
|--------|------|-------------|--------|
| GET | `/health` | Public | Liveness |
| POST | `/api/register` | Public | Creates user |
| POST | `/api/login` | Public | Returns JWT |
| GET | `/api/me` | JWT | Profile |
| GET | `/api/me/notification-preferences` | JWT | Read prefs |
| PUT | `/api/me/notification-preferences` | JWT | Update prefs |
| GET | `/api/bootstrap` | JWT | Shell config |
| GET | `/api/dashboard/patient` | JWT, patient | Patient metrics |
| GET | `/api/dashboard/doctor` | JWT, doctor | Doctor metrics |
| GET | `/api/dashboard/admin` | JWT, admin | Admin metrics |
| GET | `/api/admin/users` | JWT, admin | User list |
| PATCH | `/api/admin/users/:id` | JWT, admin | Patch user (e.g. active) |
| GET | `/api/admin` | JWT, admin | Role check |
| GET | `/api/doctor` | JWT, doctor | Role check |
| GET | `/api/patient` | JWT, patient | Role check |
| GET | `/api/geocode` | JWT, patient | Address search |
| GET | `/api/hospitals/near` | JWT, patient | Nearby hospitals |
| GET | `/api/hospitals/:id/departments` | JWT, patient | Departments |
| GET | `/api/doctors/search` | JWT, patient | Doctor search |
| GET | `/api/doctors/:id/slots` | JWT, patient | Open slots |
| POST | `/api/doctor/slots` | JWT, doctor | Create slot |
| DELETE | `/api/doctor/slots/:id` | JWT, doctor | Delete slot |
| POST | `/api/appointments/book-slot` | JWT, patient | Book from slot |
| POST | `/api/appointments` | JWT, patient | Create appointment |
| PATCH | `/api/appointments/:id/cancel` | JWT, patient | Cancel |
| PATCH | `/api/appointments/:id/request-reschedule` | JWT, patient | Reschedule request |
| GET | `/api/appointments/patient` | JWT, patient | List |
| GET | `/api/appointments/doctor` | JWT, doctor | List |
| GET | `/api/appointments/:id/activity` | JWT | Activity log |
| GET | `/api/appointments/:id` | JWT | Detail |
| GET | `/api/doctors` | JWT, patient | Directory |
| PATCH | `/api/appointments/:id/status` | JWT, doctor | Status update |
| * | `/api/messages/*` | JWT, patient **or** doctor | Threads, messages, read, unread, limits |

---

## 9. Optional: E2E (Cypress)

Not a substitute for **unit** tests in the rubric, but good proof of **integrated** behavior.

```powershell
cd frontend
npm run cypress:run
```

Requires API + `ng serve` per `docs/MANUAL-TESTING.md`.

---

## 10. Notes for graders

- **Commits:** Credit is tied to **individual commits** — verify authors on `s3-feature-updates-healthonyx` and merged feature branches.
- **Tests:** Default `go test ./...` should pass; frontend **12** unit tests as listed. Integration tests may **skip** unless env vars are set (that is expected).
- **Single source:** This file path is **`Sprint3.md`** at the repository root; GitHub renders Markdown on the web UI.

---

*Last updated for detailed submission. Replace video placeholders and teammate names before recording.*

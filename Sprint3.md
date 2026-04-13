# Sprint 3 — Submission Document

**Course submission format:** Upload this repository link and your **narrated video** link(s) where instructed. If the portal allows only one URL, paste the **GitHub** link in the main field and add **additional video links in the submission comments** (as allowed by your instructor).

---

## Repository (GitHub)

**Primary link:** [https://github.com/Abhinavasai/Healthonyx](https://github.com/Abhinavasai/Healthonyx)

**Integration branch for Sprint 3 feature work:** `s3-feature-updates-healthonyx` (individual feature branches are listed under `docs/BRANCHES.md` and `docs/sprint3-assigned-features.md`).

---

## Video presentation (requirements checklist)

Prepare a **narrated screen recording** that satisfies the assignment rubric:

| Requirement | What to show |
|--------------|----------------|
| **Split narration** | Divide the recording so **each team member narrates a distinct segment** (intro handoff, backend/API, frontend, testing & docs, wrap-up, etc.). |
| **New functionality** | Live demo of Sprint 3 features: appointment detail + activity timeline, dashboards (patient/doctor/admin as applicable), find-care (location → hospitals → departments → doctors), messaging, availability, admin user management, notification preferences — match what your team shipped on the branch above. |
| **All unit tests** | Terminal or IDE: run **frontend** and **backend** unit tests and show **green output** (include **Sprint 2** suites still in the repo — same commands below). |
| **This document** | Briefly scroll through **Sprint3.md** or summarize its sections on camera. |
| **Sprint 3 work** | Use the **“Work completed in Sprint 3”** section below as your script outline. |
| **Frontend unit test list** | Show the **Frontend unit tests** table (below) or run `ng test` with verbose output. |
| **Backend unit test list** | Show the **Backend unit tests** table (below) or `go test -v ./...`. |
| **Backend API documentation** | Open Postman collections and/or repo docs (see **API documentation**). |
| **Individual commits** | Mention that grading may use **commit history**; each contributor should have **visible commits** on the team branch. |

**Placeholders for your video URLs** (paste real links when you upload):

- **Full presentation (or Part 1):** `https://` *(add after upload)*  
- **Part 2 / supplemental:** `https://` *(optional — or use submission comments for multiple links)*  

---

## How to reproduce unit test results (for recording)

**Backend** (from repository root):

```powershell
cd backend
go test ./... -count=1
```

Optional **integration** tests (requires PostgreSQL and `DATABASE_URL` — they **skip** unless enabled):

```powershell
$env:DATABASE_URL="postgres://USER:PASSWORD@localhost:5432/healthonyx?sslmode=disable"
$env:RUN_AUTH_ME_INTEGRATION="1"
$env:RUN_APPOINTMENT_ACTIVITY_INTEGRATION="1"
go test ./handlers -v -count=1 -run "TestMe_reloadsFromDatabase|TestAppointmentActivity_flow"
```

**Frontend** (uses `ChromeHeadlessNoSandbox` via `frontend/karma.conf.js`):

```powershell
cd frontend
npx ng test --no-watch --browsers=ChromeHeadlessNoSandbox
```

**Expected:** backend `go test ./...` **passes**; frontend reports **12/12** tests passing (current suite).

---

## Work completed in Sprint 3

Summarize for the grader what the team implemented this sprint. Adjust bullets to match **your** merges and demos:

- **Appointment detail API:** `GET /api/appointments/:id` with patient/doctor authorization and consistent JSON shape.
- **Appointment activity:** persisted audit trail; `GET /api/appointments/:id/activity` with same access rules as detail; activities on approve/reject and other lifecycle actions as implemented.
- **Patient & doctor UIs:** routes `/patient/appointments/:id` and `/doctor/appointments/:id` with detail views, status display, and activity timeline.
- **Lifecycle & operations:** cancel, reschedule request, complete, no-show, and related flows where implemented (patient/doctor/admin as applicable).
- **Dashboards:** patient/doctor summary metrics; admin overview and user list with active/inactive toggles (as implemented on your branch).
- **Find care:** geocoded search, nearby hospitals, department filter tied to hospital, doctor search with geo + filters.
- **Notification preferences:** API + UI for saving preferences (as implemented).
- **Quality / docs:** hardened Postman collections (`docs/Healthonyx-API-Sprint3-Implemented.postman_collection.json`, `docs/Healthonyx-API-Complete-Regression.postman_collection.json`), manual testing guide (`docs/MANUAL-TESTING.md`), Karma configuration for reliable CI/local unit runs.

---

## Frontend unit tests

**Location:** `frontend/src/**/*.spec.ts`  
**Command:** `npx ng test --no-watch --browsers=ChromeHeadlessNoSandbox`  
**Count:** **12** tests (current codebase).

| File | Test description |
|------|-------------------|
| `patient-find-care.component.spec.ts` | shows an error if locationQuery is empty on search |
| `patient-find-care.component.spec.ts` | runs geocodeSearch and updates lat/lng + label for a single hit |
| `patient-find-care.component.spec.ts` | pickGeocodeSuggestion updates locationQuery and search point |
| `patient-find-care.component.spec.ts` | sets permission denied error when geolocation fails |
| `patient-find-care.component.spec.ts` | loadHospitals calls API with current lat/lng/radius and sets hospitals |
| `patient-find-care.component.spec.ts` | loadDoctors sets doctors from API response |
| `geo-booking.service.spec.ts` | geocodeSearch should call /api/geocode with q and limit |
| `geo-booking.service.spec.ts` | hospitalsNear should call /api/hospitals/near with lat/lng/radius_km |
| `messaging.service.spec.ts` | lists threads and updates unread from thread rows |
| `messaging.service.spec.ts` | refreshUnread uses /api/messages/unread |
| `messaging.service.spec.ts` | sendMessage posts to thread messages |
| `messaging.service.spec.ts` | createThread posts peer and body |

*Sprint 2 geo/messaging-related specs remain part of this suite; Angular runs them together.*

---

## Backend unit tests

**Command:** `go test ./...` from `backend/` (repository root: `go test ./backend/...` if needed).  
**Packages:** `handlers`, `models` (other packages have no `_test.go` files).

| File | Test function |
|------|----------------|
| `handlers/auth_me_test.go` | `TestMe_UnauthorizedWithoutClaims` |
| `handlers/auth_me_integration_test.go` | `TestMe_reloadsFromDatabase` *(integration — skipped without `RUN_AUTH_ME_INTEGRATION=1` + `DATABASE_URL`)* |
| `handlers/bootstrap_test.go` | `TestBootstrap_UnauthorizedWithoutClaims` |
| `handlers/bootstrap_test.go` | `TestBootstrap_OKWithClaims` |
| `handlers/geocode_handler_test.go` | `TestGeocodeSearch_MissingQ` |
| `handlers/geocode_handler_test.go` | `TestGeocodeSearch_ParsesUpstreamResults` |
| `handlers/geo_booking_test.go` | `TestHaversineKm_ZeroDistance` |
| `handlers/geo_booking_test.go` | `TestParseLatLngRadius_ValidDefaults` |
| `handlers/geo_booking_test.go` | `TestParseLatLngRadius_MissingLat` |
| `handlers/geo_booking_test.go` | `TestParseLatLngRadius_OutOfBounds` |
| `handlers/messaging_test.go` | `TestPreviewText_ShortUnchanged` |
| `handlers/messaging_test.go` | `TestPreviewText_TruncatesLong` |
| `handlers/messaging_test.go` | `TestValidPatientDoctorPair` |
| `handlers/messaging_test.go` | `TestComposeLimits_OK` |
| `handlers/messaging_test.go` | `TestComposeLimits_Unauthorized` |
| `handlers/messaging_test.go` | `TestUnreadTotal_Unauthorized` |
| `handlers/messaging_test.go` | `TestListMessages_InvalidThreadUUID` |
| `handlers/messaging_test.go` | `TestSendMessage_InvalidJSON` |
| `handlers/messaging_test.go` | `TestCreateThread_InvalidPeerUUID` |
| `handlers/appointments_activity_integration_test.go` | `TestAppointmentActivity_flow` *(integration — skipped without `RUN_APPOINTMENT_ACTIVITY_INTEGRATION=1` + `DATABASE_URL`)* |
| `models/appointment_activity_test.go` | `TestAppointmentActivity_JSONFields` |

---

## Updated backend API documentation

| Artifact | Path / description |
|----------|---------------------|
| **Sprint 3 Postman collection** | `docs/Healthonyx-API-Sprint3-Implemented.postman_collection.json` — import into Postman; set `baseUrl` to `http://localhost:8080`. |
| **Full regression Postman collection** | `docs/Healthonyx-API-Complete-Regression.postman_collection.json` — broader API coverage. |
| **Sprint 2 geo Postman (reference)** | `docs/Healthonyx-API-Sprint2.postman_collection.json` |
| **Manual testing guide** | `docs/MANUAL-TESTING.md` — local runbook and feature checks. |
| **Test cases & E2E matrix** | `docs/TEST-CASES-Sprint3-Implemented.md`, `docs/TEST-CASES-Frontend-Backend-E2E-Complete.md`, `docs/TEST-SETUP.md` |
| **Geo / booking notes** | `docs/geo-specialty-booking.md`, `docs/geo-specialty-booking-frontend.md` |

Run collections via Postman GUI or CLI, for example:

```powershell
npx newman run docs/Healthonyx-API-Sprint3-Implemented.postman_collection.json
```

---

## Optional: E2E (not a substitute for unit tests in the rubric)

```powershell
cd frontend
npm run cypress:run
```

Requires backend + frontend dev servers per `docs/MANUAL-TESTING.md`.

---

## Notes for graders

- **Commit history:** Work should appear as individual commits on team members’ branches merged into `s3-feature-updates-healthonyx` (or your course’s required branch).  
- **Video:** Demonstrate **working software**, **test output**, and **documentation** as required above.  
- This file is maintained in-repo at **`Sprint3.md`** (repository root).

---

*Last updated to match repository layout and scripts; replace video placeholders before submitting.*

# Sprint 3 Implemented Features - Test Cases

Scope: test only features implemented up to current Sprint 1/2/3 delivery.

## Environment

- Backend: `http://localhost:8080`
- Frontend: `http://localhost:4200`
- DB: local Postgres (`healthonyx`)
- Seed:
  - `go run ./backend/cmd/seed`
  - `go run ./backend/cmd/seed_testdata`

## Frontend Cases

| ID | Area | Steps | Expected |
|---|---|---|---|
| FE-01 | Auth Login | Open `/login`, login as `patient@healthonyx.demo` / `patient123` | Redirect to patient area, shell visible |
| FE-02 | Account Settings | Go to `/patient/settings` | Email/role/user-id card loads from `/api/me` |
| FE-03 | Find Care Geocode | In `/patient/find-care`, search "Gainesville, FL" | Geocode result appears; can continue flow |
| FE-04 | Find Nearby Hospitals | Click nearby hospitals after location search | Hospital cards/list rendered |
| FE-05 | Find Nearby Doctors | Click doctor search after location search | Doctor list rendered |
| FE-06 | Patient Appointments | Open `/patient/appointments`, create appointment | Request appears in patient list |
| FE-07 | Patient Appointment Detail | Open an appointment detail route | Detail page renders with status/reason/ids |
| FE-08 | Doctor Queue | Login as doctor and open `/doctor/appointments` | Queue renders and actions enabled for pending |
| FE-09 | Activity Timeline | Open detail with activity | Activity list loads; actor email shown when API provides it |
| FE-10 | Messaging Inbox | Open `/patient/messages` or `/doctor/messages` | Thread list + message pane visible |
| FE-11 | Messaging Compose Limit | Enter > 8000 chars in compose box | Counter turns error state; send disabled |
| FE-12 | Mobile Nav | Set mobile viewport, open shell menu | Drawer opens, closes via backdrop/link/Escape |

## Backend API Cases

| ID | Area | Endpoint(s) | Expected |
|---|---|---|---|
| BE-01 | Health | `GET /health` | `200` and `{status:"ok"}` |
| BE-02 | Register | `POST /api/register` | `201` with `token` + `user` |
| BE-03 | Login | `POST /api/login` | `200` with `token` + `user` |
| BE-04 | Profile | `GET /api/me` | `200`, DB-authoritative email/role |
| BE-05 | Bootstrap | `GET /api/bootstrap` | `200`, includes `shell_mobile_breakpoint_px` |
| BE-06 | Role Guards | `GET /api/admin|doctor|patient` | Correct role passes, wrong role blocked |
| BE-07 | Doctor Directory | `GET /api/doctors` (patient JWT) | Returns doctor options |
| BE-08 | Appointment Create/List | `POST /api/appointments`, `GET /api/appointments/patient|doctor` | Appointment persisted and listed |
| BE-09 | Appointment Status | `PATCH /api/appointments/:id/status` (doctor JWT) | Status updated |
| BE-10 | Activity Log | `GET /api/appointments/:id/activity` | Ordered activity rows; includes `actor_email` when available |
| BE-11 | Messaging Threads | `GET/POST /api/messages/threads` | Thread create + list works |
| BE-12 | Messaging Messages | `GET /api/messages/threads/:id`, `POST /api/messages/threads/:id/messages` | Messages list + send works |
| BE-13 | Messaging Unread | `GET /api/messages/unread`, `POST /api/messages/threads/:id/read` | Count updates correctly |
| BE-14 | Messaging Limits | `GET /api/messages/compose-limits` | Returns max runes metadata |

## E2E Scenario Cases

| ID | Scenario | Steps | Expected |
|---|---|---|---|
| E2E-01 | Signup + Login | Register new patient + login | Auth works, token usable |
| E2E-02 | Appointment lifecycle | Patient creates appointment, doctor approves | Both sides see updated status |
| E2E-03 | Activity trace | After approval, fetch activity | `status_changed` row present; actor metadata present |
| E2E-04 | Messaging thread | Patient opens thread to doctor, doctor replies | Both users can fetch thread messages |
| E2E-05 | Dashboard/settings sanity | User visits account/settings and shell nav | Profile renders, nav is stable |
| E2E-06 | Geo flow | Patient searches location, views hospitals/doctors | Flow succeeds without auth breaks |

## Recommended Run Order

1. Frontend: `ng build` -> `ng test`
2. Backend: `go test ./...` (+ integration tests)
3. E2E: start backend/frontend, seed DB, run Cypress + API smoke sweep


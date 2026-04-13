# Complete Regression Test Cases (Frontend, Backend, E2E)

Scope: all implemented features up to current Sprint 3 work.

## Actual Full Test Run Status (this session)

- Frontend unit tests: `12/12` passing
- Frontend production build: passing (existing non-blocking warnings only)
- Backend tests: `go test ./...` passing
- Cypress E2E full suite: `6 specs`, `9 tests`, all passing

## Test Environment

- Backend: `http://localhost:8080`
- Frontend: `http://localhost:4200`
- DB: PostgreSQL
- Seed commands:
  - `go run ./backend/cmd/seed`
  - `go run ./backend/cmd/seed_testdata`

## Frontend Test Cases

| ID | Area | Steps | Expected |
|---|---|---|---|
| FE-01 | Register/Login | Register or login with demo user | Redirect to role area with shell loaded |
| FE-02 | Account settings | Open `/patient/settings` or `/doctor/settings` | `/api/me` data shown (email/role/id) |
| FE-03 | Find care search | Search location and load hospitals/doctors | Lists render without auth redirects |
| FE-04 | Department selection | Select department in find-care dropdown | Specialty filter updates doctor search |
| FE-05 | Patient appointments | Create appointment from patient screen | Request appears in patient list |
| FE-06 | Doctor queue | Approve/reject pending request | Status updates and list refreshes |
| FE-07 | Appointment detail + activity | Open detail page and timeline | Activity rows render, actor info visible |
| FE-08 | Messaging inbox/thread | Create/open thread, send message | New message appears in thread |
| FE-09 | Messaging limits | Exceed max body size | Send disabled/error shown |
| FE-10 | Mobile nav | Open/close sidebar in mobile viewport | Toggle/backdrop/link-close all work |

## Backend API Test Cases

| ID | Area | Endpoint(s) | Expected |
|---|---|---|---|
| BE-01 | Health | `GET /health` | `200` and `{status:"ok"}` |
| BE-02 | Auth | `POST /api/register`, `POST /api/login`, `GET /api/me` | Valid JWT + profile contract |
| BE-03 | Role guards | `GET /api/admin|doctor|patient` | Role-scoped access enforced |
| BE-04 | Bootstrap | `GET /api/bootstrap` | App config payload present |
| BE-05 | Geo | `GET /api/geocode`, `/api/hospitals/near`, `/api/doctors/search` | Valid geo/hospital/doctor responses |
| BE-06 | Slots | `GET /api/doctors/:id/slots`, `GET /api/doctor/slots`, `POST/DELETE /api/doctor/slots` | Slot create/list/delete works |
| BE-07 | Appointment lifecycle | `POST /api/appointments`, `GET patient/doctor`, `PATCH status`, `GET by id` | Data persists and status transitions work |
| BE-08 | Activity log | `GET /api/appointments/:id/activity` | Ordered activity with actor fields |
| BE-09 | Patient documents | `GET/POST /api/documents`, `GET /api/documents/:id/download` | Upload/list/download works with ownership |
| BE-10 | Messaging | `/api/messages/*` (threads/messages/read/unread/compose-limits) | Threads and unread contract work |

## E2E Scenario Cases

| ID | Scenario | Steps | Expected |
|---|---|---|---|
| E2E-01 | Login + account | Login as patient, open settings | Profile page visible and populated |
| E2E-02 | Find-care flow | Search location then fetch hospitals/doctors | Geo flow succeeds end-to-end |
| E2E-03 | Appointment activity | Doctor and patient detail checks | Activity timeline states validated |
| E2E-04 | Messaging core | Open messages as patient/doctor | Inbox and thread render/send flow works |
| E2E-05 | Compose limits | Enter over-limit message | Limit enforcement validated |
| E2E-06 | Mobile navigation | Use mobile viewport nav interactions | Sidebar behavior stable |


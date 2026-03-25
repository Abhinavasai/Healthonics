# Sprint 3 — assigned features (stacked branches)

Integration base: `dev` (`b4524cf` at time of planning). Work is sequenced so each branch builds on the previous; merge in order into `dev` or rebase the next branch after each merge.

| Order | Assignee | Branch | Scope |
| ----- | -------- | ------ | ----- |
| 1 | Kaushik | `s3/kaushik-backend-appointment-detail-api` | Backend: single-appointment API |
| 2 | Pavan | `s3/pavan-frontend-patient-appointment-detail` | Frontend: patient detail screen |
| 3 | Pavan | `s3/pavan-frontend-doctor-appointment-detail` | Frontend: doctor detail screen |
| 4 | Rohith | `s3/rohith-backend-appointment-activity-log` | Backend: activity log + listing |

---

## 1. Kaushik — `GET /api/appointments/:id`

**Goal:** Let authenticated patients load their own appointment by id, and doctors load appointments assigned to them.

**Acceptance criteria**

- `GET /api/appointments/:id` requires JWT.
- Patient may only read rows where `patient_id` matches the caller.
- Doctor may only read rows where `doctor_id` matches the caller.
- Other roles receive 403; unknown id or wrong party receives 404 (no id enumeration).
- Response body is the existing `Appointment` JSON shape.

---

## 2. Pavan — Patient appointment detail (frontend)

**Goal:** From “My Requests”, open a dedicated detail view that uses Kaushik’s endpoint.

**Acceptance criteria**

- Route under `/patient/appointments/:id` (guarded for patients).
- Shows scheduled time, status, reason, doctor id (or label), with navigation back to the list.
- Uses `AppointmentsService.getById` (or equivalent) calling `GET /api/appointments/:id`.
- List entries link into the detail route.

---

## 3. Pavan — Doctor appointment detail (frontend)

**Goal:** Mirror the patient experience for doctors reviewing a single request.

**Acceptance criteria**

- Route under `/doctor/appointments/:id` (guarded for doctors).
- Shows scheduled time, status, reason, patient id, with back navigation to the queue.
- Reuses the same backend detail endpoint with the doctor JWT.

---

## 4. Rohith — Appointment activity log (backend)

**Goal:** Persist an audit trail when a doctor approves/rejects, and expose it over the API.

**Acceptance criteria**

- Migration adds `appointment_activities` (or equivalent) with FK to `appointments`.
- Successful `PATCH /api/appointments/:id/status` inserts an activity row (actor = doctor, action + optional detail).
- `GET /api/appointments/:id/activity` returns activities for that appointment; same access rules as `GET /api/appointments/:id`.

---

## Merge order

1. `s3/kaushik-backend-appointment-detail-api`
2. `s3/pavan-frontend-patient-appointment-detail`
3. `s3/pavan-frontend-doctor-appointment-detail`
4. `s3/rohith-backend-appointment-activity-log`

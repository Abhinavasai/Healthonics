# Sprint 1 – Submission Summary

**Team:** Healthonyx AI Collective  
**Sprint:** Sprint 1 (Foundation – Authentication, Role-Based Access, Placeholder Dashboards)  
**Repository:** [https://github.com/Abhinavasai/Healthonyx](https://github.com/Abhinavasai/Healthonyx)  
**Branch for Sprint 1:** `dev` (integration of all feature branches)

---

## 1. User Stories

Sprint 1 user stories are defined in BDD format in `docs/user-stories-sprint1.md`. Summary:

| ID | User Story | Summary |
|----|------------|---------|
| **US-0a** | DB Schema + Migrations | Backend: `users` table, migrations on startup, optional seed |
| **US-0b** | Password Hashing + JWT | Backend: bcrypt, JWT with user id/email/role, 401/403 |
| **US-0c** | CORS + Environment Config | Backend: .env (DATABASE_URL, JWT_SECRET, PORT, CORS), GET /health |
| **US-0d** | App Shell + Role-Based Navigation | Frontend: header/sidebar, role-specific menus, redirect to role dashboard |
| **US-1** | User Registration | New user can register (email, password, role) and be redirected to role dashboard |
| **US-2** | User Login | Registered user can log in and be redirected to role dashboard |
| **US-3** | Role-Based Access and Protected Routes | Frontend: guards/interceptor; Backend: JWT + RBAC middleware; wrong role → 403 |
| **US-4** | View Placeholder Dashboards | Patient / Doctor / Admin placeholder dashboards |
| **US-5** | Logout | Logout clears session and redirects to login |

Sub-tasks (e.g. 1a/1b, 2a/2b, 3a/3b) are listed in `docs/user-stories-sprint1-split.md`.

---

## 2. What Issues the Team Planned to Address

The team planned to address **all** of the above user stories and their acceptance criteria. In terms of GitHub/work items:

- **Backend:** US-0a (DB + migrations), US-0b (password hashing + JWT), US-0c (CORS + config), US-1b (register API), US-2b (login API), US-3b (auth + RBAC middleware), GET /api/me and role-protected endpoints (/api/patient, /api/doctor, /api/admin).
- **Frontend:** US-0d (app shell + role-based nav), US-1a (registration page), US-2a (login page), US-3a (AuthGuard, RoleGuard), US-3a-i (HTTP interceptor to attach JWT), US-4 (placeholder dashboards), US-5 (logout in app shell).

**Definition of Done (planned):**

- **Frontend:** Register and login pages working; token storage; protected and role-based routes; placeholder dashboards; logout clears session and redirects.
- **Backend:** Register and login APIs; JWT middleware; RBAC (403 for wrong role); CORS for Angular; demo via Postman/CLI.

---

## 3. Which Ones Were Successfully Completed

All planned Sprint 1 user stories were **successfully completed** and integrated into the `dev` branch.

### Backend (completed)

| Item | Status | Implementation |
|------|--------|----------------|
| US-0a: DB schema + migrations | ✅ Done | `users` table, migrations on startup, `cmd/seed` for demo users |
| US-0b: Password hashing + JWT | ✅ Done | bcrypt, JWT with id/email/role, 401/403 |
| US-0c: CORS + env config | ✅ Done | `.env` (DATABASE_URL, JWT_SECRET, PORT, CORS_ORIGINS), CORS middleware, GET /health |
| US-1b: Registration API | ✅ Done | POST /api/register, validation, returns JWT |
| US-2b: Login API | ✅ Done | POST /api/login, returns JWT + user |
| US-3b: Auth + RBAC middleware | ✅ Done | RequireAuth(), RequireRole(), GET /api/me, /api/patient, /api/doctor, /api/admin |

### Frontend (completed)

| Item | Status | Implementation |
|------|--------|----------------|
| US-0d: App shell + role-based nav | ✅ Done | AppShellComponent with header/sidebar, role-based menus, logout button |
| US-1a: Registration page | ✅ Done | Register component, form validation, API call, redirect to role dashboard |
| US-2a: Login page | ✅ Done | Login component, form, API call, redirect to role dashboard |
| US-3a-i: HTTP interceptor | ✅ Done | Auth interceptor attaches `Authorization: Bearer <token>`; 401 triggers logout |
| US-3a: AuthGuard + RoleGuard | ✅ Done | authGuard, roleGuard; unauthenticated → login; wrong role → blocked |
| US-4: Placeholder dashboards | ✅ Done | Patient / Doctor / Admin placeholders under app shell |
| US-5: Logout | ✅ Done | Logout button in app shell; clears token and redirects to login |

Feature branches merged into `dev` include, among others: `feature/US-1-user-registration`, `feature/US-2-user-login`, `feature/US-3-role-based-access`, `feature/US-0d-app-shell`, `feature/US-4-placeholder-dashboards`, `feature/US-5-logout`, `feature/US-0c-cors`, and backend auth (e.g. `feature/US-6-api-auth` or equivalent).

---

## 4. Which Ones Didn’t (and Why)

**All planned Sprint 1 user stories were completed.** There are no *incomplete* user stories for the scope the team committed to.

The following were explicitly **out of scope** or **deferred** (as per `docs/sprint1-implementation-plan.md` and `user-stories-sprint1.md`), not “failed”:

- **Optional backend:** `patient_profile` and `doctor_profile` tables (and fullName in registration) were marked optional for Sprint 1; the team delivered the required `users` table and registration/login. These can be added in a later sprint.
- **Optional frontend:** Dashboards calling backend placeholder endpoints (e.g. GET /api/patient) to show dynamic “Coming soon” text were nice-to-have; current dashboards are static placeholders. No impact on Demo Target.
- **Out of scope for Sprint 1:** File uploads, prescriptions, reminders, appointments, messaging, AI summaries. Sprint 1 scope was **identity + access + skeleton** only.

So: **we are addressing all planned Sprint 1 items.** Nothing was dropped due to failure; optional/deferred items are documented and can be scheduled for later sprints.

---


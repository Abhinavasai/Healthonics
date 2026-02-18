# Sprint 1 Implementation Plan

Definition of Done, backlog, step-by-step workflow, and status for Healthonyx Sprint 1.

---

## Sprint 1 Goal (Definition of Done)

By Feb 18 you should be able to demo:

### Frontend (Angular)

- [ ] Register screen → creates account (patient/doctor selection optional)
- [ ] Login screen → stores token
- [ ] Protected routes → blocked if not logged in
- [ ] Role-based routes → patient can't access doctor pages (and vice versa)
- [ ] Placeholder dashboards load (Patient / Doctor / Admin)
- [ ] Logout clears session and redirects

### Backend (Go + SQL)

- [ ] Register creates user + role (+ profile in DB)
- [ ] Login returns JWT
- [ ] JWT middleware protects routes
- [ ] RBAC middleware returns 403 when role mismatch
- [ ] Protected dashboard endpoints return simple JSON
- [ ] CORS enabled so Angular can call API
- [ ] Postman/curl demo works

---

## Current Status (Implementation vs Plan)

### Backend

| Plan Item | Status | Implementation |
|-----------|--------|----------------|
| Project setup (gin, pgx, jwt, bcrypt) | ✅ Done | `backend/` structure |
| GET /health | ✅ Done | `main.go` |
| DB schema (users) | ✅ Done | `db/db.go` – users table |
| patient_profile, doctor_profile | ❌ Not done | Only `users` table |
| Migrations on startup | ✅ Done | `db.Migrate()` |
| Seed demo users | ✅ Done | `cmd/seed` |
| POST /auth/register → POST /api/register | ✅ Done | `handlers/auth.go` |
| POST /auth/login → POST /api/login | ✅ Done | Returns JWT + user |
| Password hashing (bcrypt) | ✅ Done | `bcrypt.GenerateFromPassword` |
| JWT middleware | ✅ Done | `RequireAuth()` |
| RBAC middleware | ✅ Done | `RequireRole()` |
| GET /me | ✅ Done | `GET /api/me` |
| Role-protected dashboard APIs | ✅ Done | `/api/patient`, `/api/doctor`, `/api/admin` |
| .env config | ✅ Done | `config/config.go`, godotenv |
| CORS | ❌ Not done | Using Angular proxy in dev |

### Frontend

| Plan Item | Status | Implementation |
|-----------|--------|----------------|
| Routes: /register, /login | ✅ Done | `app.routes.ts` |
| Routes: /patient, /doctor, /admin | ✅ Done | Flat paths (not /patient/dashboard) |
| AuthService | ✅ Done | `auth.service.ts` |
| register(), login(), logout() | ✅ Done | |
| Token storage (localStorage) | ✅ Done | |
| Register component | ✅ Done | `register.component` |
| Login component | ❌ Placeholder only | "Coming in US-2" – needs real form |
| HTTP Interceptor (attach JWT) | ❌ Not done | |
| AuthGuard | ❌ Not done | |
| RoleGuard | ❌ Not done | |
| App shell (toolbar, sidebar, logout) | ❌ Not done | |
| Placeholder dashboards | ✅ Basic | "Coming soon" text |
| Dashboards call backend | ❌ Not done | Static placeholder only |

---

## API Route Mapping

| Plan | Current | Notes |
|------|---------|-------|
| POST /auth/register | POST /api/register | ✅ |
| POST /auth/login | POST /api/login | ✅ |
| GET /me | GET /api/me | ✅ |
| GET /patient/dashboard | GET /api/patient | Different path; returns `{message}` |
| GET /doctor/dashboard | GET /api/doctor | Same |
| GET /admin/dashboard | GET /api/admin | Same |

---

## Implementation Workflow

### A) Backend Plan

#### A1) Project Setup ✅

- `backend/` with main.go, handlers, db, config
- gin, pgx, jwt, bcrypt, godotenv
- GET /health

#### A2) Database Schema ⚠️ Partial

**Current:** `users` (id, email, password_hash, role, created_at)

**Plan adds:** `patient_profile`, `doctor_profile` (optional for Sprint 1)

#### A3) Registration API ✅

POST /api/register – validate, hash, insert, return JWT

#### A4) Login API ✅

POST /api/login – validate, bcrypt compare, issue JWT

#### A5) Middleware: Auth + RBAC ✅

RequireAuth() + RequireRole()

#### A6) Protected Dashboard APIs ✅

GET /api/me, /api/patient, /api/doctor, /api/admin

#### A7) CORS + Config ❌ CORS missing

Add CORS middleware for http://localhost:4200

---

### B) Frontend Plan

#### B1) App Skeleton + Routing ⚠️ Partial

- Routes exist; no shared layout/shell

#### B2) Auth Service + Token Storage ✅

AuthService with register, login, logout, getToken, getUser, isLoggedIn

#### B3) HTTP Interceptor ❌

Add interceptor to attach `Authorization: Bearer <token>`

#### B4) Guards ❌

AuthGuard + RoleGuard

#### B5) UI Pages ⚠️ Partial

- Register ✅
- Login ❌ (placeholder only)
- Dashboards ⚠️ (basic; no data from backend)
- App shell with logout ❌

---

## Integration Checklists

### Backend Checklist

- [x] Migrations run with 1 command (on startup)
- [x] Register inserts user
- [ ] Register inserts patient_profile/doctor_profile *(optional)*
- [x] Login returns valid JWT
- [x] Middleware blocks missing/invalid token (401)
- [x] RBAC blocks wrong role (403)
- [ ] CORS allows Angular origin
- [x] /health works

### Frontend Checklist

- [x] Register persists token and user on success
- [ ] Login persists token and role
- [ ] Interceptor attaches JWT to API calls
- [ ] AuthGuard blocks dashboards when logged out
- [ ] RoleGuard blocks wrong dashboard
- [ ] Logout clears token and redirects (logic exists; no UI button)
- [ ] App shell with logout button
- [ ] Dashboards display placeholder data from backend *(optional)*

---

## What's Left to Build (in order)

| # | ID | Item | Depends On |
|---|-----|------|------------|
| 1 | 0c | CORS middleware | — |
| 2 | 2a | Login page (form + API + redirect) | 2b ✅ |
| 3 | 3a-i | HTTP interceptor (attach JWT) | 2a |
| 4 | 3a | AuthGuard + RoleGuard | 2a, 3a-i |
| 5 | 0d | App shell (header, sidebar, logout) | 3a |
| 6 | 5 | Logout (button in shell) | 0d |

### Nice-to-have

- patient_profile / doctor_profile tables + fullName in registration
- Dashboards fetch from backend placeholder endpoints
- Postman collection or curl examples in README

---

## Sprint 1 Demo Script

### Frontend (2 members)

1. Register as patient
2. Login as patient → patient dashboard loads
3. Manually open /doctor → blocked/redirected
4. Logout → redirected to login

### Backend (Postman/curl)

1. POST /api/register – create user
2. POST /api/login – get JWT
3. GET /api/me with Bearer token → 200
4. GET /api/doctor with doctor token → 200
5. GET /api/doctor with patient token → 403

---

## What NOT to Build in Sprint 1

- File uploads
- Prescriptions / reminders
- Appointments
- Messaging
- AI summaries / KB monitor

Sprint 1 = **identity + access + skeleton**.

---

## GitHub Backlog (for issue creation)

Aligned with work item IDs. Create in this order:

| # | ID | Title | Type | AC Summary |
|---|-----|-------|------|------------|
| 1 | 0c | Add CORS middleware for Angular origin | Backend | CORS allows http://localhost:4200 |
| 2 | 2a | Login page – form + API + redirect | Frontend | Email, password, submit → token → redirect to dashboard |
| 3 | 3a-i | HTTP interceptor – attach JWT | Frontend | All /api/* requests get Authorization: Bearer token |
| 4 | 3a | AuthGuard + RoleGuard | Frontend | No token → /login; wrong role → redirect to correct dashboard |
| 5 | 0d | App shell – header, sidebar, logout | Frontend | Shared layout with role-based nav and logout button |
| 6 | 5 | Logout – wire button to AuthService | Frontend | Button in shell calls logout(), clears token, redirects |

---

## Document Index

| Document | Purpose |
|----------|---------|
| `user-stories-sprint1.md` | Full BDD user stories (US-0 through US-5) |
| `user-stories-sprint1-split.md` | Assignable work items + implementation order + backlog |
| `sprint1-implementation-plan.md` | Definition of Done, status, checklists, remaining work |

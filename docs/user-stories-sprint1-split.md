# Sprint 1 Work Items (Split) – Assignable

Sprint 1 split into frontend/backend sub-tasks for clear assignment.

**Assignees (example):** 1a/2a Abhinava | 1b/2b Siddani | 3a/4/5 Pavan | 0a-0c/3b Rohith

**See also:** `user-stories-sprint1.md` (full BDD stories) | `sprint1-implementation-plan.md` (status, checklists)

---

## Quick Reference – Implementation Order

```
Backend:  0a → 0b → 1b → 2b → 3b → 0c
Frontend: 1a → 2a → 3a-i → 3a → 0d → 4 → 5
```

**Remaining (6 items):** 0c (CORS), 2a (Login page), 3a-i (Interceptor), 3a (Guards), 0d (App shell), 5 (Logout)

---

## Backend Foundation (do first)

### US-0a: DB Schema + Migrations

**As a** backend developer  
**I want** a proper DB schema with migrations  
**So that** registration and login can persist real user data.

**Acceptance Criteria:**
- Tables: `users` (id, email, password_hash, role, created_at)
- Migrations run on startup
- Optional: seed demo users

---

### US-0b: Password Hashing + JWT

**As a** backend API  
**I want** to hash passwords (bcrypt) and issue JWT tokens  
**So that** credentials are secure and clients can authenticate.

**Acceptance Criteria:**
- Passwords hashed with bcrypt
- JWT includes user id, email, role
- 401 / 403 responses for auth failures

---

### US-0c: CORS + Environment Config

**As a** backend API  
**I want** CORS enabled and config from .env  
**So that** Angular can call the API and deployment is flexible.

**Acceptance Criteria:**
- .env loading (DATABASE_URL, JWT_SECRET, PORT)
- CORS allows Angular origin (e.g. http://localhost:4200)
- GET /health endpoint

---

## Frontend Foundation

### US-0d: App Shell + Role-Based Navigation

**As a** logged-in user  
**I want** a header/sidebar layout with menus that match my role  
**So that** the app has a consistent, professional shell.

**Acceptance Criteria:**
- Header/sidebar wrapping dashboard routes
- Menus differ for Patient vs Doctor vs Admin
- After login, redirect to role dashboard

---

## US-1a: Registration – Frontend

**As a** new user  
**I want to** see a registration form and get validation feedback  
**So that** I can register and reach my dashboard.

### Acceptance Criteria (Given/When/Then)

- **Given** I am on the registration page  
  **When** I enter valid email, password (6+ chars), and select a role  
  **Then** the form submits to the backend and, on success, I am logged in and redirected to my role dashboard.

- **Given** I am on the registration page  
  **When** I submit with an email that is already registered  
  **Then** I see "Email already registered" and remain on the page.

- **Given** I am on the registration page  
  **When** I submit with invalid data (empty email, bad format, short password)  
  **Then** I see validation errors and the form is not submitted.

---

## US-1b: Registration – Backend

**As a** backend API  
**I want to** accept registration requests and persist users with role  
**So that** new users can have an account.

### Acceptance Criteria (Given/When/Then)

- **Given** a valid payload (email, password, role)  
  **When** the request is processed  
  **Then** the user is stored and the API returns success (and JWT).

- **Given** a payload with an email that already exists  
  **When** the request is processed  
  **Then** the API responds with 409 (or 400) "email already registered".

- **Given** invalid data (weak password, invalid role)  
  **When** the request is processed  
  **Then** the API responds with validation errors and does not create a user.

---

## US-2a: Login – Frontend

**As a** registered user  
**I want to** use a login form to sign in  
**So that** I can access my dashboard.

### Acceptance Criteria (Given/When/Then)

- **Given** I am on the login page  
  **When** I enter correct email and password  
  **Then** I am logged in and redirected to my role dashboard.

- **Given** I am on the login page  
  **When** I enter incorrect email or password  
  **Then** I see "Invalid email or password" and remain on the page.

- **Given** I am already logged in  
  **When** I navigate to the login page  
  **Then** I am redirected to my dashboard.

---

## US-2b: Login – Backend

**As a** backend API  
**I want to** validate credentials and issue JWT  
**So that** clients can authenticate.

### Acceptance Criteria (Given/When/Then)

- **Given** correct email and password  
  **When** the request is processed  
  **Then** the API returns JWT with user id, email, role.

- **Given** incorrect email or password  
  **When** the request is processed  
  **Then** the API responds with 401 Unauthorized.

---

## US-3a-i: HTTP Interceptor (attach JWT)

**As a** frontend application  
**I want** all API requests to include the JWT automatically  
**So that** protected endpoints receive the token.

**Acceptance Criteria:**
- Interceptor attaches `Authorization: Bearer <token>` to `/api/*` requests
- On 401 response, redirect to /login

---

## US-3a: Protected Routes – Frontend (AuthGuard + RoleGuard)

**As a** logged-in user  
**I want** the frontend to show only routes for my role  
**So that** I cannot navigate to unauthorized areas.

### Acceptance Criteria (Given/When/Then)

- **Given** I am logged in as a patient  
  **When** I try to open /doctor or /admin  
  **Then** I am redirected to login or forbidden.

- **Given** I am logged in as a doctor  
  **When** I try to open /patient or /admin  
  **Then** I am redirected to login or forbidden.

- **Given** I am not logged in  
  **When** I try to open /patient, /doctor, or /admin  
  **Then** I am redirected to the login page.

---

## US-3b: Auth + RBAC Middleware (Backend)

**Combined:** JWT validation + role-based access control. *(Replaces separate US-6.)*

**As a** backend API  
**I want** JWT validation and RBAC on protected endpoints  
**So that** only authorized clients and roles can access sensitive data.

### Acceptance Criteria (Given/When/Then)

- **Given** a request to GET /api/me without Authorization header  
  **When** the request is processed  
  **Then** the server responds with 401 Unauthorized.

- **Given** a request to GET /api/me with valid Bearer JWT  
  **When** the request is processed  
  **Then** the server responds with 200 and user profile (id, email, role).

- **Given** a request to a role-protected route (e.g. /api/doctor/*) with JWT of wrong role  
  **When** the request is processed  
  **Then** the server responds with 403 Forbidden.

---

## US-4: View Placeholder Dashboards

**As a** logged-in patient, doctor, or admin  
**I want to** see a dedicated dashboard placeholder for my role  
**So that** I have a clear entry point for future features.

### Acceptance Criteria (Given/When/Then)

- **Given** I am logged in as a patient  
  **When** I am on the patient dashboard  
  **Then** I see a "Patient Dashboard" placeholder (e.g. "Coming soon").

- **Given** I am logged in as a doctor  
  **When** I am on the doctor dashboard  
  **Then** I see a "Doctor Dashboard" placeholder.

- **Given** I am logged in as an admin  
  **When** I am on the admin dashboard  
  **Then** I see an "Admin Dashboard" placeholder.

---

## US-5: Logout

**As a** logged-in user  
**I want to** log out from the application  
**So that** my session is cleared.

### Acceptance Criteria (Given/When/Then)

- **Given** I am logged in and on a dashboard with the main layout (sidebar)  
  **When** I click Logout  
  **Then** I am logged out, token cleared, redirected to login.

- **Given** I have logged out  
  **When** I try to open /patient, /doctor, or /admin  
  **Then** I am redirected to the login page.

---

## Implementation Order (with dependencies)

Work items must be done in this order (or parallel where noted) so dependencies are met.

### Phase 1: Backend foundation (do first)

| Order | ID | Work Item | Depends On |
|-------|-----|-----------|------------|
| 1 | 0a | DB schema + migrations | — |
| 2 | 0b | Password hashing + JWT | 0a |
| 3 | 1b | Registration – Backend | 0a, 0b |
| 4 | 2b | Login – Backend | 0a, 0b |
| 5 | 3b | Auth + RBAC middleware | 0b |
| 6 | 0c | CORS + env config | — (can do anytime) |

### Phase 2: Frontend auth flow

| Order | ID | Work Item | Depends On |
|-------|-----|-----------|------------|
| 7 | 1a | Registration – Frontend | 1b |
| 8 | 2a | Login – Frontend | 2b |
| 9 | 3a-i | HTTP interceptor (attach JWT) | 2a |
| 10 | 3a | AuthGuard + RoleGuard | 2a, 3a-i |

### Phase 3: Layout and polish

| Order | ID | Work Item | Depends On |
|-------|-----|-----------|------------|
| 11 | 0d | App shell + role nav | 3a |
| 12 | 4 | Placeholder dashboards | 3a, 0d |
| 13 | 5 | Logout | 0d, 3a |

---

## Suggested Sprint 1 Backlog

| ID | Work Item | Type | Owner | Status |
|----|-----------|------|-------|--------|
| 0a | DB schema + migrations | Backend | Backend 1 | ✅ Done |
| 0b | Password hashing + JWT | Backend | Backend 1 | ✅ Done |
| 0c | CORS + env config | Backend | Backend 1 | ❌ |
| 0d | App shell + role nav | Frontend | Frontend 1 | ❌ |
| 1a | Registration – Frontend | Frontend | Abhinava | ✅ Done |
| 1b | Registration – Backend | Backend | Siddani | ✅ Done |
| 2a | Login – Frontend | Frontend | Abhinava | ❌ |
| 2b | Login – Backend | Backend | Siddani | ✅ Done |
| 3a | Protected routes (AuthGuard, RoleGuard) | Frontend | Pavan | ❌ |
| 3a-i | HTTP interceptor | Frontend | Pavan | ❌ |
| 3b | Auth + RBAC middleware | Backend | Rohith | ✅ Done |
| 4 | Placeholder dashboards | Frontend | Pavan | ✅ Done (basic) |
| 5 | Logout | Frontend | Pavan | ❌ |

---

*Sprint 1 work items for Healthonyx AI Collective. Update assignees and status as needed.*

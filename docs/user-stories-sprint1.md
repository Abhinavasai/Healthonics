# Sprint 1 User Stories (BDD)

Sprint 1 focuses on project foundation: authentication, role-based access, and placeholder dashboards.

---

## US-0: Backend Foundation (Must-have before auth flows)

### US-0a: DB Schema + Migrations

**As a** backend developer  
**I want** a proper DB schema with migrations and seed data  
**So that** registration and login can persist real user data.

**Acceptance Criteria:**
- Tables: `users` (id, email, password_hash, role, created_at)
- Migrations run on startup (or via migration tool)
- Optional: seed default roles/demo users

---

### US-0b: Password Hashing + JWT Implementation

**As a** backend API  
**I want** to hash passwords and issue JWT tokens  
**So that** credentials are secure and clients can authenticate.

**Acceptance Criteria:**
- Passwords hashed with bcrypt (or argon2)
- JWT access token with user id, email, role
- Standard error responses: 401 Unauthorized, 403 Forbidden

---

### US-0c: CORS + Environment Config

**As a** backend API  
**I want** CORS enabled and config loaded from environment  
**So that** the Angular frontend can call the API and deployment is flexible.

**Acceptance Criteria:**
- `.env` loading for DATABASE_URL, JWT_SECRET, PORT
- CORS middleware allows Angular origin (e.g. http://localhost:4200)
- GET /health endpoint for health checks

---

## US-0d: App Shell + Role-Based Navigation (Frontend Foundation)

**As a** logged-in user  
**I want** a consistent layout (header/sidebar) with navigation that matches my role  
**So that** the app feels professional and I can reach my dashboards.

**Acceptance Criteria:**
- Header/sidebar layout wrapping dashboard routes
- Menus differ for Patient vs Doctor vs Admin
- After login, redirect to role-specific dashboard

---

## US-1: User Registration

**As a** new user  
**I want to** register with email, password, and role (patient, doctor, or admin)  
**So that** I can access the Healthonyx platform with the appropriate permissions.

### Acceptance Criteria (Given/When/Then)

- **Given** I am on the registration page  
  **When** I enter a valid email, password (at least 6 characters), and select a role  
  **Then** I am registered and logged in, and I am redirected to my role-specific dashboard (patient, doctor, or admin).

- **Given** I am on the registration page  
  **When** I submit with an email that is already registered  
  **Then** I see an error message that the email is already registered and I remain on the registration page.

- **Given** I am on the registration page  
  **When** I submit with invalid data (e.g. empty email, invalid email format, or password too short)  
  **Then** I see validation errors and the form is not submitted.

---

## US-2: User Login

**As a** registered user  
**I want to** log in with my email and password  
**So that** I can access my dashboard and use the platform.

### Acceptance Criteria (Given/When/Then)

- **Given** I am on the login page  
  **When** I enter correct email and password  
  **Then** I am logged in and redirected to my role-specific dashboard (patient → /patient, doctor → /doctor, admin → /admin).

- **Given** I am on the login page  
  **When** I enter incorrect email or password  
  **Then** I see an error message (e.g. "Invalid email or password") and I remain on the login page.

- **Given** I am logged in  
  **When** I navigate to the login page  
  **Then** I can be redirected to my dashboard (or the login page is not accessible when already authenticated, depending on product rule).

---

## US-3: Role-Based Access and Protected Routes

**As a** logged-in user  
**I want** the application to show only the dashboard and features for my role  
**So that** I cannot access areas I am not authorized to use.

### Frontend (US-3a)

- **Given** I am logged in as a patient  
  **When** I try to open the doctor or admin dashboard URL  
  **Then** I am redirected to login (or a forbidden page) and cannot see the doctor or admin content.

- **Given** I am logged in as a doctor  
  **When** I try to open the patient or admin dashboard URL  
  **Then** I am redirected to login (or forbidden) and cannot see that content.

- **Given** I am not logged in  
  **When** I try to open any of /patient, /doctor, or /admin  
  **Then** I am redirected to the login page.

### Backend (US-3b: Auth + RBAC Middleware)

**Combined:** JWT validation + role-based access control.

- **Given** a request to GET /api/me without an Authorization header  
  **When** the request is processed  
  **Then** the server responds with 401 Unauthorized.

- **Given** a request to GET /api/me with a valid Bearer JWT  
  **When** the request is processed  
  **Then** the server responds with 200 and the current user's profile (e.g. id, email, role).

- **Given** a request to a role-protected route (e.g. /api/doctor/*) with a JWT that has a different role  
  **When** the request is processed  
  **Then** the server responds with 403 Forbidden.

---

## US-4: View Placeholder Dashboards

**As a** logged-in patient, doctor, or admin  
**I want to** see a dedicated dashboard placeholder for my role  
**So that** I have a clear entry point for future features (appointments, prescriptions, etc.).

### Acceptance Criteria (Given/When/Then)

- **Given** I am logged in as a patient  
  **When** I am on the patient dashboard  
  **Then** I see a "Patient Dashboard" placeholder (e.g. "Coming soon" with a short description).

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
**So that** my session is cleared and others cannot use my account on the same device.

### Acceptance Criteria (Given/When/Then)

- **Given** I am logged in and on a dashboard page with the main layout (sidebar)  
  **When** I click Logout  
  **Then** I am logged out, my token is cleared, and I am redirected to the login page.

- **Given** I have logged out  
  **When** I try to open /patient, /doctor, or /admin  
  **Then** I am redirected to the login page.

---

## Implementation Order

**Phase 1 (Backend):** 0a → 0b → 1b → 2b → 3b → 0c  
**Phase 2 (Frontend auth):** 1a → 2a → 3a-i (Interceptor) → 3a (AuthGuard, RoleGuard)  
**Phase 3 (Layout):** 0d (App shell) → 4 (Dashboards) → 5 (Logout)

*Full order with dependencies: `user-stories-sprint1-split.md`*

---

## Sprint 1 Demo Target

By end of Sprint 1, you should be able to demo:

**Frontend**
- Register page works (form validation, API call)
- Login page works
- After login → redirected to role dashboard
- Protected routes block access if not logged in
- Logout clears session + redirects to login

**Backend (Postman/CLI)**
- POST /api/register creates user + returns JWT
- POST /api/login returns JWT
- GET /api/me returns user info if JWT present
- /api/doctor/* blocked for patients (403)
- /api/patient/* blocked for doctors (403)

---

*These user stories cover the scope of Sprint 1 (Foundation) and can be used for BDD scenarios and acceptance testing.*

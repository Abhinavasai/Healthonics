# Sprint 1 User Stories (Split) – Assignable

Sprint 1 split into frontend/backend sub-tasks for clear assignment.  
**Assignees:** 1a/2a Abhinava | 1b/2b Siddani | 3a/4/5 Pavan | 3b/6 Thandava Sai Rohith Achanta

---

## US-1a: Registration – Frontend

**As a** new user  
**I want to** see a registration form and get feedback on validation  
**So that** I can register with email, password, and role (patient, doctor, or admin) and reach my dashboard.

### Acceptance Criteria (Given/When/Then)

- **Given** I am on the registration page  
  **When** I enter a valid email, password (at least 6 characters), and select a role  
  **Then** the form submits to the backend and, on success, I am logged in and redirected to my role-specific dashboard (patient, doctor, or admin).

- **Given** I am on the registration page  
  **When** I submit with an email that is already registered  
  **Then** I see an error message that the email is already registered and I remain on the registration page.

- **Given** I am on the registration page  
  **When** I submit with invalid data (e.g. empty email, invalid email format, or password too short)  
  **Then** I see validation errors and the form is not submitted.

---

## US-1b: Registration – Backend

**As a** backend API  
**I want to** accept registration requests and persist users with role  
**So that** new users can have an account and be authenticated.

### Acceptance Criteria (Given/When/Then)

- **Given** a valid registration payload (email, password, role)  
  **When** the request is processed  
  **Then** the user is stored, and the API returns a success response (and optionally a token or session).

- **Given** a registration payload with an email that already exists  
  **When** the request is processed  
  **Then** the API responds with an error (e.g. 409 or 400) indicating the email is already registered.

- **Given** a registration payload with invalid data (e.g. weak password, invalid role)  
  **When** the request is processed  
  **Then** the API responds with validation errors and does not create a user.

---

## US-2a: Login – Frontend

**As a** registered user  
**I want to** use a login form to sign in  
**So that** I can access my dashboard and the platform.

### Acceptance Criteria (Given/When/Then)

- **Given** I am on the login page  
  **When** I enter correct email and password and submit  
  **Then** I am logged in and redirected to my role-specific dashboard (patient → /patient, doctor → /doctor, admin → /admin).

- **Given** I am on the login page  
  **When** I enter incorrect email or password  
  **Then** I see an error message (e.g. "Invalid email or password") and I remain on the login page.

- **Given** I am already logged in  
  **When** I navigate to the login page  
  **Then** I am redirected to my dashboard (or the login page is not accessible when authenticated).

---

## US-2b: Login – Backend

**As a** backend API  
**I want to** validate credentials and issue a session or JWT  
**So that** clients can authenticate and access protected resources.

### Acceptance Criteria (Given/When/Then)

- **Given** a login request with correct email and password  
  **When** the request is processed  
  **Then** the API returns a success response and a token (e.g. JWT) that includes the user id, email, and role.

- **Given** a login request with incorrect email or password  
  **When** the request is processed  
  **Then** the API responds with 401 Unauthorized (or equivalent) and no token.

---

## US-3a: Protected Routes – Frontend

**As a** logged-in user  
**I want** the frontend to show only routes and dashboards for my role  
**So that** I cannot navigate to areas I am not authorized to use.

### Acceptance Criteria (Given/When/Then)

- **Given** I am logged in as a patient  
  **When** I try to open the doctor or admin dashboard URL  
  **Then** I am redirected to login or a forbidden page and cannot see the doctor or admin content.

- **Given** I am logged in as a doctor  
  **When** I try to open the patient or admin dashboard URL  
  **Then** I am redirected to login or forbidden and cannot see that content.

- **Given** I am not logged in  
  **When** I try to open any of /patient, /doctor, or /admin  
  **Then** I am redirected to the login page.

---

## US-3b: Role Checks – Backend

**As a** backend API  
**I want** to enforce role-based access on protected endpoints  
**So that** only authorized roles can perform certain actions.

### Acceptance Criteria (Given/When/Then)

- **Given** a request to a role-protected endpoint with a JWT that has the wrong role  
  **When** the request is processed  
  **Then** the server responds with 403 Forbidden.

- **Given** a request to a role-protected endpoint with a valid JWT and correct role  
  **When** the request is processed  
  **Then** the server processes the request and returns the appropriate response.

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

## US-6: API Authentication (Backend)

**As a** backend API  
**I want** to accept JWT on protected endpoints and enforce role-based access  
**So that** only authorized clients can access sensitive data.

### Acceptance Criteria (Given/When/Then)

- **Given** a request to GET /api/me without an Authorization header  
  **When** the request is processed  
  **Then** the server responds with 401 Unauthorized.

- **Given** a request to GET /api/me with a valid Bearer JWT  
  **When** the request is processed  
  **Then** the server responds with 200 and the current user's profile (e.g. id, email, role).

- **Given** a request to a role-protected route with a JWT that has a different role  
  **When** the request is processed  
  **Then** the server responds with 403 Forbidden.

---

*Sprint 1 split stories for Healthonyx AI Collective. Assign: 1a/2a Abhinava Sai Tirunagari | 1b/2b Siddani Kaushik Bhargav | 3a/4/5 Pavan Karthik Chilla | 3b/6 Thandava Sai Rohith Achanta.*

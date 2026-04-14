# Healthonyx

## Project Description
Healthonyx is an AI-enabled healthcare monitoring and medical record management platform designed to simplify healthcare workflows. The system provides role-based access for doctors, patients, and administrators to securely manage medical records, appointments, prescriptions, and health alerts. By automating routine processes and offering intelligent insights, Healthonyx reduces manual workload while improving efficiency and quality of care.

## Team Name
Healthonyx AI Collective

## Team Members
- Abhinava Sai Tirunagari – Frontend Engineer
- Pavan Karthik Chilla – Frontend Engineer
- Siddani Kaushik Bhargav – Backend Engineer
- Rohith Achanta – Backend Engineer

---

## Development Setup

### Prerequisites
- **Go** 1.21+ (backend)
- **Node.js** 20+ and **npm** (frontend)
- **PostgreSQL** (e.g. local install or [ElephantSQL](https://www.elephantsql.com/) free tier)

### Backend

1. From the repo root, go to the backend directory:
   ```bash
   cd backend
   ```
2. Copy the example env file and set your database URL and JWT secret:
   ```bash
   cp .env.example .env
   ```
   Edit `.env` and set at least:
   - `DATABASE_URL` – e.g. `postgres://user:pass@localhost:5432/healthonyx?sslmode=disable`
   - `JWT_SECRET` – a secure random string for production
3. Install dependencies and run:
   ```bash
   go mod download
   go run .
   ```
   The API will listen on `http://localhost:8080`. Migrations run on startup.

4. **(Optional)** Seed demo users (admin, doctor, patient):
   ```bash
   go run ./cmd/seed
   ```
   Default demo logins (see seed output): e.g. `admin@healthonyx.demo` / `admin123`, `doctor@healthonyx.demo` / `doctor123`, `patient@healthonyx.demo` / `patient123`.

5. Run backend unit tests (Go):
   ```bash
   go test ./...
   ```

### Frontend

1. From the repo root, go to the frontend directory:
   ```bash
   cd frontend
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Start the dev server (with proxy to the backend at `http://localhost:8080`):
   ```bash
   npm start
   ```
   The app will be at `http://localhost:4200`. Use **Login** or **Register**; after login you are redirected to the role-specific dashboard (patient, doctor, or admin).

4. Run frontend unit tests (Angular/Jasmine via Karma, headless, bundled Chromium via Puppeteer):
   ```bash
   cd frontend
   npm install
   npm run test:ci
   ```
   This sets `CI=true` so Karma uses `ChromeHeadlessCI` (sandbox-friendly flags) and `CHROME_BIN` from `puppeteer` when system Chrome is missing.

5. Run Cypress E2E (starts API + `ng serve`, then runs all headless specs):
   ```bash
   # Terminal A — database (once)
   docker compose up -d
   # Ensure backend/.env has DATABASE_URL matching docker-compose (see below)
   cd backend && go run ./cmd/seed && cd ..
   # Terminal B — from frontend (needs Go on PATH)
   cd frontend
   npm install
   npm run cypress:e2e
   ```
   The script starts the API on **port 28180** (avoids clashing with a dev API on 8080), proxies `/api` from Angular on **4300** to that API, and passes matching `API_URL` to Cypress. Prerequisites: PostgreSQL reachable, `backend/.env` with `DATABASE_URL` and `JWT_SECRET`, seed run at least once, and **ports 28180 and 4300 free** (change `PORT` / `proxy.e2e.conf.json` / `cy:run` together if you need a different port). For manual runs against the usual stack (API 8080, app 4200): start API + `npm start`, then `npx cypress run --headless --spec "cypress/e2e/find-care.cy.js"` (uses default `API_URL` from `cypress.config.js`).

6. Production build:
   ```bash
   npm run build
   ```

### Running both together
- Terminal 1: `cd backend && go run .`
- Terminal 2: `cd frontend && npm start`
- Open `http://localhost:4200` and log in or register.

Note: if you just cloned the repo and the database is empty, run the backend once (migrations happen on startup) and then run:
`cd backend && go run ./cmd/seed`

#### Recommended local run order
1. Start PostgreSQL (Docker or local) and make sure `backend/.env` has `DATABASE_URL`.
2. Start backend: `cd backend && go run .`
3. (Optional) Seed demo users: `go run ./cmd/seed`
4. Start frontend: `cd frontend && npm start`

### Docker (optional) for PostgreSQL
From the repo root:
```bash
docker compose up -d
```
Then set `DATABASE_URL=postgres://healthonyx:healthonyx@localhost:5432/healthonyx?sslmode=disable` in `backend/.env`.

Alternatively, a one-off container:
```bash
docker run -d --name healthonyx-db -e POSTGRES_USER=healthonyx -e POSTGRES_PASSWORD=healthonyx -e POSTGRES_DB=healthonyx -p 5432:5432 postgres:16
```

---

## CI
GitHub Actions runs on push/PR to `main` (or `master`): backend `go build` and `go test`, frontend `npm run test` and `npm run build`. See [.github/workflows/ci.yml](.github/workflows/ci.yml).

---

## Sprint 1 documentation

| Doc | Purpose |
|-----|---------|
| [user-stories-sprint1.md](docs/user-stories-sprint1.md) | BDD user stories (US-0 through US-5) |
| [user-stories-sprint1-split.md](docs/user-stories-sprint1-split.md) | Assignable work items, implementation order, backlog |
| [sprint1-implementation-plan.md](docs/sprint1-implementation-plan.md) | Definition of Done, status, checklists |
| [TEST-SETUP.md](docs/TEST-SETUP.md) | How to run and test |

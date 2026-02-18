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

5. Run tests:
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

4. Run tests:
   ```bash
   npm run test
   ```
5. Production build:
   ```bash
   npm run build
   ```

### Running both together
- Terminal 1: `cd backend && go run .`
- Terminal 2: `cd frontend && npm start`
- Open `http://localhost:4200` and log in or register.

### Docker (optional) for PostgreSQL only
If you prefer to run Postgres in Docker:
```bash
docker run -d --name healthonyx-db -e POSTGRES_USER=healthonyx -e POSTGRES_PASSWORD=healthonyx -e POSTGRES_DB=healthonyx -p 5432:5432 postgres:16
```
Then set `DATABASE_URL=postgres://healthonyx:healthonyx@localhost:5432/healthonyx?sslmode=disable` in `backend/.env`.

---

## CI
GitHub Actions runs on push/PR to `main` (or `master`): backend `go build` and `go test`, frontend `npm run test` and `npm run build`. See [.github/workflows/ci.yml](.github/workflows/ci.yml).

---

## Sprint 1 documentation
BDD user stories for Sprint 1 are in [docs/user-stories-sprint1.md](docs/user-stories-sprint1.md).

# Test Setup – Sprint 1 (Current Scope)

Quick guide to run and test the current implementation.

**Done:** 0a, 0b, 1a, 1b, 2b, 3b, 4 (basic) | **Remaining:** 0c, 2a, 3a-i, 3a, 0d, 5

## Prerequisites

- **Go** 1.21+
- **Node.js** 20+
- **Docker** (for PostgreSQL), OR a local PostgreSQL instance

## 1. Start PostgreSQL

### Option A: Docker (recommended)

```powershell
docker run -d --name healthonyx-db -e POSTGRES_USER=healthonyx -e POSTGRES_PASSWORD=healthonyx -e POSTGRES_DB=healthonyx -p 5432:5432 postgres:16
```

If the container already exists: `docker start healthonyx-db`

### Option B: Local PostgreSQL

Create a database named `healthonyx` and set `DATABASE_URL` in `backend/.env`:

```
DATABASE_URL=postgres://YOUR_USER:YOUR_PASSWORD@localhost:5432/healthonyx?sslmode=disable
```

## 2. Backend (.env)

The repo includes `backend/.env` for local testing. If missing, copy from example:

```powershell
cd backend
copy .env.example .env
```

Edit `.env` if your database URL differs.

## 3. Run Backend

```powershell
cd backend
go mod download
go run .
```

API listens on `http://localhost:8080`. Migrations run on startup.

## 4. Seed Demo Users (optional)

In another terminal:

```powershell
cd backend
go run ./cmd/seed
```

Demo logins: `admin@healthonyx.demo` / `admin123`, `doctor@healthonyx.demo` / `doctor123`, `patient@healthonyx.demo` / `patient123`

## 5. Run Frontend

```powershell
cd frontend
npm install
npm start
```

App runs at `http://localhost:4200`. The dev server proxies `/api` to the backend.

## 6. Test Registration (US-1a + US-1b)

1. Open **http://localhost:4200**
2. Use the **Register** form
3. Enter email, password (6+ chars), and select role (patient/doctor/admin)
4. Click **Register**
5. You should be redirected to the role dashboard (e.g. `/patient`)

### Test cases

| Test | Expected |
|------|----------|
| Valid registration | Redirected to role dashboard |
| Duplicate email | "Email already registered" error |
| Empty email | Validation error |
| Invalid email format | Validation error |
| Password &lt; 6 chars | Validation error |

# Sprint 1 Test Checklist

**Test from:** `test_p2` (integration branch – has all completed work)

**Branch → Work mapping:** See [BRANCHES.md](BRANCHES.md)

---

## Prerequisites

- Go 1.21+, Node.js 20+, PostgreSQL (or Docker)
- On `test_p2`: `git checkout test_p2`

---

## 1. Start PostgreSQL

```powershell
docker start healthonyx-db
# Or if not created: docker run -d --name healthonyx-db -e POSTGRES_USER=healthonyx -e POSTGRES_PASSWORD=healthonyx -e POSTGRES_DB=healthonyx -p 5432:5432 postgres:16
```

---

## 2. Backend

```powershell
cd backend
go mod tidy
go run .
```

Optional seed:
```powershell
go run ./cmd/seed
```

---

## 3. Frontend

```powershell
cd frontend
npm install
npm start
```

Open http://localhost:4200

---

## 4. Test Cases

| # | Test | Steps | Expected |
|---|------|-------|----------|
| 1 | Register | Register with new email, 6+ char password, role | Redirect to role dashboard |
| 2 | Duplicate email | Register again with same email | "Email already registered" |
| 3 | Register validation | Submit with invalid email or short password | Validation errors |
| 4 | Login | Use seeded user (e.g. admin@healthonyx.demo / admin123) | Redirect to dashboard |
| 5 | Login invalid | Wrong email/password | "Invalid email or password" |
| 6 | API (Postman) | POST /api/login → GET /api/me with Bearer token | 200 + user profile |

---

## 5. What’s Not Implemented Yet

- Route protection (anyone can open /patient, /doctor, /admin)
- HTTP interceptor (JWT not auto-attached)
- App shell (sidebar, logout button)

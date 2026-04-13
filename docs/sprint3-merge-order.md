# Sprint 3 — merge order into `dev`

Follow this sequence **feature by feature**. For each step:

1. Merge the **frontend** branch into `dev` (PR).
2. Rebase (or merge) the matching **backend** branch onto latest `origin/dev`, then merge into `dev` (PR).

Do **not** skip steps; order reduces dependency and merge conflicts.

| Step | Frontend → `dev` | Then backend → `dev` |
|------|------------------|----------------------|
| 1 | `s3-feature09-frontend-karthik` | `s3-feature09-backend-kaushik` |
| 2 | `s3-feature03-frontend-karthik` | `s3-feature03-backend-kaushik` |
| 3 | `s3-feature04-frontend-abhinav` | `s3-feature04-backend-rohith` |
| 4 | `s3-feature05-frontend-karthik` | `s3-feature05-backend-kaushik` |
| 5 | `s3-feature06-frontend-abhinav` | `s3-feature06-backend-rohith` |
| 6 | `s3-feature07-frontend-abhinav` | `s3-feature07-backend-kaushik` |
| 7 | `s3-feature08-frontend-karthik` | `s3-feature08-backend-rohith` |
| 8 | `s3-feature10-frontend-abhinav` | `s3-feature10-backend-rohith` |
| 9 | `s3-feature01-frontend-abhinav` | `s3-feature01-backend-kaushik` |
| 10 | `s3-feature02-frontend-karthik` | `s3-feature02-backend-rohith` |

**Backend branch before PR:** `git fetch origin && git checkout <backend-branch> && git rebase origin/dev` (resolve conflicts, then `git push --force-with-lease` if already pushed).

---

## Sprint 4 — merge order into `dev` (after Sprint 3 is fully merged)

Sprint 4 branches are **full-stack** (each branch usually includes both backend and frontend). Merge **one branch at a time** into `dev` in this order; it reduces overlap on `backend/db/db.go`, `backend/main.go`, `app.routes.ts`, and `app-shell.component.ts`.

| Step | Branch → `dev` | Notes |
|------|------------------|--------|
| 1 | `feature/s4-abhinav-patient-files` | Patient files table, uploads, `/patient/my-files` |
| 2 | `feature/s4-abhinav-prescriptions` | Prescriptions table + patient UI |
| 3 | `feature/s4-kaushik-notifications` | Notifications table + inbox |
| 4 | `feature/s4-kaushik-appointment-lifecycle` | Status/cancel/activity extensions |
| 5 | `feature/s4-karthik-dashboards` | Dashboard summaries + patient/doctor dashboards |
| 6 | `feature/s4-karthik-admin-audit` | `audit_logs`, `/admin/audit` |
| 7 | `feature/s4-rohith-appointment-comments` | `appointment_comments`; keep route order `.../comments` before `.../:id` in `main.go` |
| 8 | `feature/s4-rohith-knowledge-base` | `knowledge_docs`, `/admin/knowledge`; union with audit (one admin shell, both children) |

Before each PR: `git fetch origin && git checkout <branch> && git rebase origin/dev` (or merge `origin/dev` in), resolve conflicts, `go test ./...` (backend), `ng build` (frontend).

---

## Combined S3 + S4 — single sequence (target: `dev`)

| Global step | What to merge |
|-------------|----------------|
| 1–10 | Sprint 3: same as the first table (each step: **frontend PR**, then **backend PR**). |
| 11 | `feature/s4-abhinav-patient-files` |
| 12 | `feature/s4-abhinav-prescriptions` |
| 13 | `feature/s4-kaushik-notifications` |
| 14 | `feature/s4-kaushik-appointment-lifecycle` |
| 15 | `feature/s4-karthik-dashboards` |
| 16 | `feature/s4-karthik-admin-audit` |
| 17 | `feature/s4-rohith-appointment-comments` |
| 18 | `feature/s4-rohith-knowledge-base` |

**Rule of thumb:** do not start Sprint 4 until **both** sides of Sprint 3 step 10 are merged and `dev` is green. Optional: open PRs from `https://github.com/Abhinavasai/Healthonyx` for each `feature/s4-*` branch after push.

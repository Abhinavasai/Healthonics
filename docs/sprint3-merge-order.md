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

---

## S3.5 — integration line (full S4 stack + latest `dev` documents)

To reduce repeated merge pain, the repo maintains an **integration branch** that unions **Sprint 4 feature branches** with **`origin/dev`** (Sprint 3 partial + patient/doctor documents module). Use it for end-to-end testing and as a single PR target when `dev` is still catching up on S3 steps 4–10.

| Branch | Purpose |
|--------|---------|
| `integration/s4-complete-merge` | Canonical integration: S4 eight features merged with conflict resolution; merge `origin/dev` into it to pick up latest Sprint 3/documents work. |
| `integration/s3.5-full` | **Same tip as** `integration/s4-complete-merge` (alias for naming). Full-stack smoke: `go test ./...`, `ng build`, Cypress against API + `ng serve`. |
| `s3.5-abhinav` | Same commit as integration tip — **assignee lane** for Abhinav (patient files, prescriptions, Abhinav S3 fronts); use for reviews/CI, not a different codebase. |
| `s3.5-karthik` | Assignee lane — Karthik (dashboards, admin shell pieces, Karthik S3 fronts, messaging UI). |
| `s3.5-kaushik` | Assignee lane — Kaushik (Kaushik S3 backends, notifications, appointment lifecycle). |
| `s3.5-rohith` | Assignee lane — Rohith (Rohith S3 backends, document summary API path, comments, knowledge). |

**Remaining S3 work** (steps 4–10 backend/frontend pairs not yet in `dev`) should still land via the first table; merge or rebase them into `dev` first, then merge `dev` into `integration/s4-complete-merge` again to refresh the integration line.

**Note:** The matrix you maintain against the PDF still has gaps (delivery, KB health, AI/RAG, etc.); “~60% portal MVP” refers to integrated clinic flows, not the full PDF + AI roadmap.

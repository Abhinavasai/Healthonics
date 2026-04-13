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

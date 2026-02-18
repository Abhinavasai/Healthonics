# Branch-to-Work-Item Mapping

Each user story / work item has its own feature branch.

| Branch | Work Item | Description |
|--------|-----------|-------------|
| `feature/US-0c-cors` | 0c | CORS + env config |
| `feature/US-0d-app-shell` | 0d | App shell + role nav |
| `feature/US-1-user-registration` | 1a, 1b | Registration (FE + BE) |
| `feature/US-2-user-login` | 2a, 2b | Login (FE + BE) |
| `feature/US-3-role-based-access` | 3a, 3a-i | Protected routes, AuthGuard, RoleGuard, Interceptor |
| `feature/US-4-placeholder-dashboards` | 4 | Placeholder dashboards |
| `feature/US-5-logout` | 5 | Logout |
| `feature/US-6-api-auth` | 0a, 0b, 3b | DB schema, JWT, Auth + RBAC middleware |
| `test_p1` | — | Integration branch for combined testing |

## Workflow

1. Create/checkout branch for work item: `git checkout feature/US-XXX`
2. Implement, commit
3. Merge into `test_p1` for integration testing
4. Merge `test_p1` into `main` when Sprint 1 is complete

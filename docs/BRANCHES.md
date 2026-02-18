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
| `test_p1` | — | Legacy integration branch |
| `test_p2` | — | **Integration branch for combined testing** (current work) |

## Linear History (Rebased)

All branches are rebased onto `project-skeleton` in this order:

```
main → project-skeleton → US-6 → US-1 → US-0c → US-2 → docs → US-3 → US-0d → US-5
                                                                              ↑
                                                                          test_p2
```

Each feature branch points to its commit in this chain. Merge in this order for conflict-free integration.

## Workflow

1. Base: `project-skeleton` (structure only)
2. Create/checkout branch for work item: `git checkout feature/US-XXX`
3. Implement, commit
4. Merge into `test_p2` in order (US-6 → US-1 → US-0c → US-2 → … → US-5)
5. Merge `test_p2` into `dev` or `main` when Sprint 1 is complete

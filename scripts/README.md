# Scripts

## create-sprint-issues.ps1

Creates GitHub issues from the Sprint 1 user stories markdown file using the GitHub CLI.

### Prerequisites

- [GitHub CLI (gh)](https://cli.github.com/) installed and authenticated: `gh auth login`
- Run from the repository root or ensure `docs/user-stories-sprint1-split.md` exists relative to the repo

### Usage

```powershell
# From repo root (Healthonyx)
cd "d:\OneDrive - University of Florida\Software Engineering\Healthonyx"

# Preview only (no issues created)
.\scripts\create-sprint-issues.ps1 -DryRun

# Create issues with default label "sprint-1"
.\scripts\create-sprint-issues.ps1

# Custom labels
.\scripts\create-sprint-issues.ps1 -Label "sprint-1,user-story,frontend"

# Different stories file or repo
.\scripts\create-sprint-issues.ps1 -StoriesPath "docs/sprint2-stories.md" -Repo "owner/healthonyx"
```

### Parameters

| Parameter      | Default                            | Description                                                  |
|----------------|------------------------------------|--------------------------------------------------------------|
| `StoriesPath`  | `docs/user-stories-sprint1-split.md` | Path to the work items markdown file                       |
| `DryRun`       | false                         | Only list what would be created                              |
| `Label`        | `sprint-1`                    | Comma-separated labels for each issue                       |
| `Repo`         | current repo                  | Target repo as `owner/name`                                  |
| `Assignees`    | (none)                        | Comma-separated GitHub usernames, one per issue in order    |

### Split stories (13 work items with assignees)

The default file contains **13 work items** (0a, 0b, 0c, 0d, 1a, 1b, 2a, 2b, 3a-i, 3a, 3b, 4, 5). Pass assignees in this order:

| # | ID   | Item                     | Type    |
|---|------|--------------------------|---------|
| 1 | 0a   | DB schema + migrations   | Backend |
| 2 | 0b   | Password hashing + JWT   | Backend |
| 3 | 0c   | CORS + env config        | Backend |
| 4 | 0d   | App shell + role nav     | Frontend|
| 5 | 1a   | Registration – Frontend  | Frontend|
| 6 | 1b   | Registration – Backend   | Backend |
| 7 | 2a   | Login – Frontend         | Frontend|
| 8 | 2b   | Login – Backend          | Backend |
| 9 | 3a-i | HTTP interceptor         | Frontend|
|10 | 3a   | AuthGuard + RoleGuard    | Frontend|
|11 | 3b   | Auth + RBAC middleware   | Backend |
|12 | 4    | Placeholder dashboards   | Frontend|
|13 | 5    | Logout                   | Frontend|

Replace the placeholder usernames below with your team’s actual GitHub usernames (same order as above):

```powershell
.\scripts\create-sprint-issues.ps1 -StoriesPath "docs/user-stories-sprint1-split.md" -Repo "Abhinavasai/Healthonyx" -Assignees "Abhinavasai,SiddaniKaushik,PavanKarthik,Abhinavasai,SiddaniKaushik,PavanKarthik,PavanKarthik,PavanKarthik,RohithAchanta"
```

Use `-Label ""` to skip labels if your repo doesn’t allow creating them.

### Next steps to create issues in GitHub

1. **Install GitHub CLI** (if not installed): https://cli.github.com/  
   - Windows: `winget install GitHub.cli` or download the MSI from the site.

2. **Log in to GitHub** from a terminal:
   ```powershell
   gh auth login
   ```
   Follow the prompts (browser or token).

3. **Ensure your Healthonyx repo is on GitHub**  
   If it’s only local, create a repo on GitHub and add it as `origin`, then push.

4. **Run the script (no dry-run)** from the repo root:
   ```powershell
   cd "d:\OneDrive - University of Florida\Software Engineering\Healthonyx"
   .\scripts\create-sprint-issues.ps1
   ```
   This creates 13 issues from the split doc with label `sprint-1`. To add more labels:
   ```powershell
   .\scripts\create-sprint-issues.ps1 -Label "sprint-1,user-story"
   ```

5. **If the repo is under a different owner/name**, pass it explicitly:
   ```powershell
   .\scripts\create-sprint-issues.ps1 -Repo "YourUsername/Healthonyx"
   ```

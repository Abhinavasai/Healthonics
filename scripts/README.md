# Scripts

## create-sprint-issues.ps1

Creates GitHub issues from the Sprint 1 user stories markdown file using the GitHub CLI.

### Prerequisites

- [GitHub CLI (gh)](https://cli.github.com/) installed and authenticated: `gh auth login`
- Run from the repository root or ensure `docs/user-stories-sprint1.md` exists relative to the repo

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

| Parameter      | Default                       | Description                                                  |
|----------------|-------------------------------|--------------------------------------------------------------|
| `StoriesPath`  | `docs/user-stories-sprint1.md`| Path to the user stories markdown file                       |
| `DryRun`       | false                         | Only list what would be created                              |
| `Label`        | `sprint-1`                    | Comma-separated labels for each issue                       |
| `Repo`         | current repo                  | Target repo as `owner/name`                                  |
| `Assignees`    | (none)                        | Comma-separated GitHub usernames, one per issue in order    |

### Split stories (9 issues with assignees)

To create **9 issues** (US-1a/1b, US-2a/2b, US-3a/3b, US-4, US-5, US-6) and assign them in one go, use the split file and pass assignees in this order:

| # | Issue   | Assignee (team member)              |
|---|--------|--------------------------------------|
| 1 | US-1a  | Abhinava Sai Tirunagari (frontend)   |
| 2 | US-1b  | Siddani Kaushik Bhargav (backend)    |
| 3 | US-2a  | Abhinava Sai Tirunagari (frontend)   |
| 4 | US-2b  | Siddani Kaushik Bhargav (backend)    |
| 5 | US-3a  | Pavan Karthik Chilla (frontend)      |
| 6 | US-3b  | Thandava Sai Rohith Achanta (backend)|
| 7 | US-4   | Pavan Karthik Chilla (frontend)      |
| 8 | US-5   | Pavan Karthik Chilla (frontend)      |
| 9 | US-6   | Thandava Sai Rohith Achanta (backend)|

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
   This creates 6 issues with label `sprint-1`. To add more labels:
   ```powershell
   .\scripts\create-sprint-issues.ps1 -Label "sprint-1,user-story"
   ```

5. **If the repo is under a different owner/name**, pass it explicitly:
   ```powershell
   .\scripts\create-sprint-issues.ps1 -Repo "YourUsername/Healthonyx"
   ```

<#
.SYNOPSIS
    Creates GitHub issues from a sprint user-stories markdown file.

.DESCRIPTION
    Parses docs/user-stories-sprint1-split.md (or another file), splits by "## US-N:" sections
    (supports US-0a, US-3a-i, etc.), and creates one GitHub issue per work item using the GitHub CLI (gh).

.PARAMETER StoriesPath
    Path to the user stories markdown file. Default: docs/user-stories-sprint1-split.md (relative to repo root).

.PARAMETER DryRun
    If set, only prints what would be created; does not call gh.

.PARAMETER Label
    Comma-separated labels to add to each issue (e.g. "sprint-1,user-story"). Creates labels if missing.

.PARAMETER Repo
    Target repo as owner/name. Default: current repo from git.

.PARAMETER Assignees
    Comma-separated GitHub usernames, one per issue in order (13 items: 0a,0b,0c,0d,1a,1b,2a,2b,3a-i,3a,3b,4,5).

.EXAMPLE
    .\create-sprint-issues.ps1 -DryRun
    Preview issues from split doc without creating.

.EXAMPLE
    .\create-sprint-issues.ps1 -StoriesPath docs/user-stories-sprint1-split.md -Repo "owner/Healthonyx" -Assignees "rohith,rohith,rohith,pavan,abhinava,siddani,abhinava,siddani,pavan,pavan,rohith,pavan,pavan"
    Create 13 split issues (0a,0b,0c,0d,1a,1b,2a,2b,3a-i,3a,3b,4,5) and assign.
#>

param(
    [Parameter()]
    [string] $StoriesPath = "docs/user-stories-sprint1-split.md",

    [Parameter()]
    [switch] $DryRun,

    [Parameter()]
    [string] $Label = "sprint-1",

    [Parameter()]
    [string] $Repo = "",

    [Parameter()]
    [string] $Assignees = ""
)

$ErrorActionPreference = "Stop"

# Resolve path relative to script directory (repo root if script is in /scripts)
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Split-Path -Parent $ScriptDir
$StoriesFullPath = if ([System.IO.Path]::IsPathRooted($StoriesPath)) { $StoriesPath } else { Join-Path $RepoRoot $StoriesPath }

if (-not (Test-Path $StoriesFullPath)) {
    Write-Error "Stories file not found: $StoriesFullPath"
    exit 1
}

# Check for gh (only when actually creating issues)
if (-not $DryRun -and -not (Get-Command gh -ErrorAction SilentlyContinue)) {
    Write-Error "GitHub CLI (gh) is not installed or not in PATH. Install from https://cli.github.com/"
    exit 1
}

$content = Get-Content -Path $StoriesFullPath -Raw
if (-not $content) {
    Write-Error "File is empty: $StoriesFullPath"
    exit 1
}

# Extract each ## US-N: or ### US-N: section (title + body until next --- ## or ### or end)
$stories = [System.Collections.ArrayList]::new()
$pattern = '(?ms)^#{2,3} (US-\d+[a-z0-9-]*:[^\r\n]+)\r?\n\r?\n(.*?)(?=\r?\n---\r?\n\r?\n#{2,3} |\r?\n\*These user stories|\r?\n\*Sprint 1 split|\r?\n\*Sprint 1 work|\z)'
$matches = [regex]::Matches($content, $pattern)
foreach ($m in $matches) {
    $title = $m.Groups[1].Value.Trim()
    if ($title -match '^US-') {
        $body = $m.Groups[2].Value.Trim()
        [void]$stories.Add([PSCustomObject]@{ Title = $title; Body = $body })
    }
}

if ($stories.Count -eq 0) {
    Write-Error "No user story sections (## US-N: or ## US-Nx: ...) found in $StoriesFullPath"
    exit 1
}

$assigneeList = @()
if ($Assignees) {
    $assigneeList = @(($Assignees -split ',\s*' | ForEach-Object { $_.Trim() }) | Where-Object { $_ })
}

Write-Host "Found $($stories.Count) user stories. DryRun=$DryRun" -ForegroundColor Cyan

$labelArgs = @()
$labelsToCreate = @()
if ($Label) {
    foreach ($l in $Label -split ',\s*') {
        $trimmed = $l.Trim()
        if ($trimmed) {
            $labelArgs += "--label", $trimmed
            $labelsToCreate += $trimmed
        }
    }
}

$repoArgs = @()
if ($Repo) {
    $repoArgs = @("--repo", $Repo)
}

# Ensure labels exist (create if missing). Non-fatal: if this fails (404, permissions), continue without labels.
if (-not $DryRun -and $labelsToCreate.Count -gt 0) {
    $prevErrAction = $ErrorActionPreference
    $ErrorActionPreference = "SilentlyContinue"
    $labelOk = $true
    foreach ($lbl in $labelsToCreate) {
        $labelCreateArgs = @("label", "create", $lbl, "--color", "0E8A16", "--description", "Sprint work item") + $repoArgs
        $null = & gh @labelCreateArgs 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Note: Could not create label '$lbl' (repo may not exist or lack permission). Continuing without labels." -ForegroundColor Yellow
            $labelArgs = @()
            $labelOk = $false
            break
        }
    }
    $ErrorActionPreference = $prevErrAction
}

$issueIndex = 0
foreach ($s in $stories) {
    Write-Host "`n--- $($s.Title) ---" -ForegroundColor Yellow
    if ($DryRun) {
        Write-Host "Body preview (first 200 chars): $($s.Body.Substring(0, [Math]::Min(200, $s.Body.Length)))..."
        $issueIndex++
        continue
    }

    $tempBody = [System.IO.Path]::GetTempFileName()
    try {
        [System.IO.File]::WriteAllText($tempBody, $s.Body, [System.Text.UTF8Encoding]::new($false))

        $procArgs = @("issue", "create", "--title", $s.Title, "--body-file", $tempBody) + $labelArgs + $repoArgs
        $result = & gh @procArgs 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Error "gh issue create failed: $result"
            exit $LASTEXITCODE
        }
        Write-Host "Created: $result" -ForegroundColor Green

        # Assign if we have an assignee for this issue index
        if ($assigneeList.Count -gt $issueIndex -and $assigneeList[$issueIndex]) {
            $assignee = $assigneeList[$issueIndex]
            if ($result -match '/issues/(\d+)$') {
                $issueNum = $matches[1]
                $editArgs = @("issue", "edit", $issueNum, "--add-assignee", $assignee) + $repoArgs
                $null = & gh @editArgs 2>&1
                if ($LASTEXITCODE -eq 0) {
                    Write-Host "Assigned to @$assignee" -ForegroundColor Gray
                }
            }
        }
    }
    finally {
        if (Test-Path $tempBody) { Remove-Item $tempBody -Force }
    }
    $issueIndex++
}

Write-Host "`nDone." -ForegroundColor Cyan

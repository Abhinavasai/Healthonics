# Healthonyx - Run for Testing (Sprint 1)
# Test from: test_p2 (integration branch)
# Run this from the repo root after starting PostgreSQL.

param(
    [switch]$Db,      # Start PostgreSQL in Docker
    [switch]$Seed     # Seed demo users after backend is running
)

$RepoRoot = Split-Path -Parent $PSScriptRoot

if ($Db) {
    Write-Host "Starting PostgreSQL in Docker..." -ForegroundColor Cyan
    docker start healthonyx-db 2>$null
    if ($LASTEXITCODE -ne 0) {
        docker run -d --name healthonyx-db -e POSTGRES_USER=healthonyx -e POSTGRES_PASSWORD=healthonyx -e POSTGRES_DB=healthonyx -p 5432:5432 postgres:16
    }
    Write-Host "Waiting for PostgreSQL to be ready..."
    Start-Sleep -Seconds 5
}

Write-Host ""
Write-Host "Ensure you're on test_p2: git checkout test_p2" -ForegroundColor Cyan
Write-Host "To test Sprint 1, open TWO terminals:" -ForegroundColor Yellow
Write-Host ""
Write-Host "Terminal 1 - Backend:" -ForegroundColor Green
Write-Host "  cd $RepoRoot\backend"
Write-Host "  go run ."
Write-Host ""
Write-Host "Terminal 2 - Frontend:" -ForegroundColor Green
Write-Host "  cd $RepoRoot\frontend"
Write-Host "  npm start"
Write-Host ""
Write-Host "Then open http://localhost:4200 and use the Register page." -ForegroundColor Cyan
Write-Host ""

if ($Seed) {
    Write-Host "Seeding demo users..." -ForegroundColor Cyan
    Set-Location "$RepoRoot\backend"
    go run ./cmd/seed
}

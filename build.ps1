# waktu-sholat build script (PowerShell) — GOTH stack
# Usage:
#   .\build.ps1 gen       — generate templ files
#   .\build.ps1 css       — compile Tailwind CSS
#   .\build.ps1 css-watch — watch Tailwind CSS
#   .\build.ps1 web       — build web binary (templ + tailwind + go build)
#   .\build.ps1 desktop   — build Wails desktop app
#   .\build.ps1 run       — run web server
#   .\build.ps1 check     — run qibla self-check
#   .\build.ps1 all       — gen + css + web

param([string]$Action = "all")

$ErrorActionPreference = "Stop"
$gobin = (go env GOPATH) + "\bin"
$env:PATH = "$gobin;$env:PATH"

switch ($Action) {
    "gen" {
        Write-Host "Generating templ files..." -ForegroundColor Cyan
        templ generate
    }
    "css" {
        Write-Host "Compiling Tailwind CSS..." -ForegroundColor Cyan
        npx @tailwindcss/cli -i input.css -o static/out.css --minify
    }
    "css-watch" {
        Write-Host "Watching Tailwind CSS..." -ForegroundColor Cyan
        npx @tailwindcss/cli -i input.css -o static/out.css --watch
    }
    "web" {
        Write-Host "Building web binary..." -ForegroundColor Cyan
        templ generate
        npx @tailwindcss/cli -i input.css -o static/out.css --minify
        go build -o waktu-sholat.exe .
        Write-Host "Done. Binary: waktu-sholat.exe" -ForegroundColor Green
    }
    "desktop" {
        Write-Host "Building desktop binary..." -ForegroundColor Cyan
        templ generate
        npx @tailwindcss/cli -i input.css -o static/out.css --minify
        go build -tags desktop -o waktu-sholat-desktop.exe .
        Write-Host "Done. Binary: waktu-sholat-desktop.exe" -ForegroundColor Green
    }
    "wails-build" {
        Write-Host "Building Wails app (wails build)..." -ForegroundColor Cyan
        templ generate
        npx @tailwindcss/cli -i input.css -o static/out.css --minify
        wails build
    }
    "run" {
        Write-Host "Running web server..." -ForegroundColor Cyan
        templ generate
        npx @tailwindcss/cli -i input.css -o static/out.css --minify
        go run . --addr=127.0.0.1:9090
    }
    "check" {
        go run . check
    }
    "all" {
        Write-Host "Full build: templ + tailwind + go build..." -ForegroundColor Cyan
        templ generate
        npx @tailwindcss/cli -i input.css -o static/out.css --minify
        go build -o waktu-sholat.exe .
        Write-Host "Done. Binary: waktu-sholat.exe" -ForegroundColor Green
    }
    default {
        Write-Host "Unknown action: $Action" -ForegroundColor Red
        Write-Host "Valid: gen, css, css-watch, web, desktop, wails-build, run, check, all"
        exit 1
    }
}

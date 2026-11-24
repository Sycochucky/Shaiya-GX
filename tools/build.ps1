# Run in PowerShell with Go and an x64 MinGW-w64 gcc on PATH.
$ErrorActionPreference = "Stop"
Set-Location (Split-Path -Parent $PSScriptRoot)
$env:CGO_ENABLED = "1"
New-Item -ItemType Directory -Force build,server/build | Out-Null
function Check-Go { if ($LASTEXITCODE -ne 0) { throw "Go command failed: $LASTEXITCODE" } }
go vet ./...
Check-Go
foreach ($program in @("shaiya", "shaiyacli", "assets")) {
    $target = if ($program -eq "shaiya") { "Shaiya GX" } else { $program }
    go build -o "build/$target.exe" "./cmd/$program"
    Check-Go
}
New-Item -ItemType Directory -Force "build/GX Studio" | Out-Null
go build -trimpath -ldflags "-H windowsgui" -o "build/GX Studio/GX Studio.exe" ./cmd/gx_studio
Check-Go
Push-Location .
try {
    & ./server/tools/build.ps1 -Gui
    Check-Go
} finally {
    Pop-Location
}
go test ./...
Check-Go

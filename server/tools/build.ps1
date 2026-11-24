# Build the headless server and its optional standalone desktop controller.
param([switch]$Gui)
$ErrorActionPreference = "Stop"
Set-Location (Split-Path -Parent $PSScriptRoot)
$previousCGO = $env:CGO_ENABLED
try {
    $env:CGO_ENABLED = "0"
    go build -o build\ ./cmd/gx_login ./cmd/gx_world ./cmd/gx_console
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    if ($Gui) {
        $env:CGO_ENABLED = "1"
        go build -trimpath -ldflags "-H windowsgui" -o build\gx_server_control.exe ./cmd/gx_server_control
    }
} finally {
    if ($null -eq $previousCGO) { Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue }
    else { $env:CGO_ENABLED = $previousCGO }
}
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

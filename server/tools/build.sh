#!/bin/sh
# Build the headless server and its optional standalone desktop controller.
set -e
if [ "$#" -gt 1 ] || { [ "$#" -eq 1 ] && [ "$1" != --gui ]; }; then
    echo 'Usage: server/tools/build.sh [--gui]' >&2
    exit 2
fi
repo_root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$repo_root"
if [ -x "$HOME/.local/opt/go-1.27.4/bin/go" ]; then
    PATH="$HOME/.local/opt/go-1.27.4/bin:$PATH"
    export PATH
fi
export GOTOOLCHAIN=local
CGO_ENABLED=0 go build -o build/ ./cmd/gx_login ./cmd/gx_world ./cmd/gx_console
if [ "${1:-}" = --gui ]; then
    CGO_ENABLED=1 go build -trimpath -o build/gx_server_control ./cmd/gx_server_control
fi

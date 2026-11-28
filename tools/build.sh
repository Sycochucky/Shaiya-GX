#!/usr/bin/env bash
# Build and check the integrated Linux client and GX Server.
set -euo pipefail
cd "$(dirname "$0")/.."
source tools/go-env.sh
mkdir -p build server/build
go vet ./...
for c in shaiya shaiyacli assets; do
  if [ "$c" = shaiya ]; then target="Shaiya GX"; else target="$c"; fi
  go build -o "build/$target" "./cmd/$c"
done
./server/tools/build.sh --gui
go test -vet=off ./...

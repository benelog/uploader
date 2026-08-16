#!/usr/bin/env bash
# Cross-compile the release binaries into dist/, one per supported platform.
# Standalone: safe to run any time; it only writes to dist/.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

rm -rf dist && mkdir -p dist
for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do
  GOOS=${target%/*}
  GOARCH=${target#*/}
  out=dist/uploader-$GOOS-$GOARCH
  [[ $GOOS == windows ]] && out=$out.exe
  GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$out" .
  echo "built $out"
done

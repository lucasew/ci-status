#!/bin/bash
set -e
mkdir build -p
echo "# Built targets" >> $GITHUB_STEP_SUMMARY
export CGO_ENABLED=0

# Build bootstrap binary
go build -v -o ci-status ./cmd/ci-status

# First pass: Create all statuses as pending/queued
go tool dist list | grep -vE 'wasm|aix|plan9|android|ios|illumos|solaris|dragonfly' | while IFS=/ read -r GOOS GOARCH; do
  ./ci-status set "Build $GOOS/$GOARCH" \
    --state pending \
    --description "Queued" || true
done

# Build each target
go tool dist list | grep -vE 'wasm|aix|plan9|android|ios|illumos|solaris|dragonfly' | while IFS=/ read -r GOOS GOARCH; do
  echo "::group::Build $GOOS/$GOARCH"
  GOOS=$GOOS GOARCH=$GOARCH ./ci-status run "Build $GOOS/$GOARCH" \
    -- \
    go build -v -o build/ci-status-$GOOS-$GOARCH ./cmd/ci-status \
    && (echo "- $GOOS/$GOARCH" >> $GITHUB_STEP_SUMMARY) || true
  echo "::endgroup::"
done

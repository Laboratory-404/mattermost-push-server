#!/usr/bin/env bash
# Build Script for Cross-Compiling Mattermost Push Proxy for Linux
set -e

mkdir -p build

echo "Building Linux x86_64 (amd64) binary..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/mattermost-push-proxy-linux-amd64 ./cmd/server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/test-push-linux-amd64 ./cmd/test-push

echo "Building Linux ARM64 binary..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o build/mattermost-push-proxy-linux-arm64 ./cmd/server
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o build/test-push-linux-arm64 ./cmd/test-push

echo "Building local native binary..."
go build -ldflags="-s -w" -o build/mattermost-push-proxy ./cmd/server
go build -ldflags="-s -w" -o build/test-push ./cmd/test-push

echo "✅ Build complete! Binaries located in ./build/"

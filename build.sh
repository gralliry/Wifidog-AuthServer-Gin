#!/bin/bash

# https://freshman.tech/snippets/go/cross-compile-go-programs/
# debug | release | test
GIN_MODE='debug'

# ----------------- #
# centos7
env GOOS=linux GOARCH=amd64 GIN_MODE="$GIN_MODE" go build -ldflags="-s -w" -o authserver

# windows11
env GOOS=windows GOARCH=amd64 GIN_MODE="$GIN_MODE" go build -ldflags="-s -w" -o authserver.exe

# darwin
env GOOS=darwin GOARCH=amd64 GIN_MODE="$GIN_MODE" go build -ldflags="-s -w" -o authserver

# openwrt
env GOOS=linux GOARCH=mipsle GIN_MODE="$GIN_MODE" GOMIPS='softfloat' go build -ldflags="-s -w" -o authserver


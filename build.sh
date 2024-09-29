#!/bin/bash

# linux | darwin | windows
GOOS='linux'
# amd64 | arm64 | mipsle
GOARCH='mipsle'
# softfloat | hardfloat
GOMIPS='softfloat'
# debug | release | test
GIN_MODE='debug'

env GOOS="$GOOS" GOARCH="$GOARCH" GOMIPS="$GOMIPS" \
  GIN_MODE="$GIN_MODE" CGO_ENABLED=1 \
  go build -ldflags="-s -w" -o bin/authserver-"$GOOS"-"$GOARCH"

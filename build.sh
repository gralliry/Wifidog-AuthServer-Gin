#!/bin/bash

# linux | darwin | windows
GOOS='linux'
# amd64 | arm64 | mipsle
# https://freshman.tech/snippets/go/cross-compile-go-programs/
GOARCH='mipsle'
# softfloat | hardfloat
GOMIPS='softfloat'
# debug | release | test
GIN_MODE='debug'
# gcc path
GCC_PATH='/root/buildroot/output/host/usr/bin/mipsel-linux-gcc'

env GOOS="$GOOS" GOARCH="$GOARCH" GOMIPS="$GOMIPS" \
  GIN_MODE="$GIN_MODE" \
  CGO_ENABLED=0 CC="$GCC_PATH" \
  go build -ldflags="-s -w" -o authserver-"$GOOS"-"$GOARCH"

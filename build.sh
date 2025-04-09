#!/bin/bash

GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 CC=clang go build -o bin/lrcAPI_v$1_darwin_arm64
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 CC="clang -arch x86_64" go build -o bin/lrcAPI_v$1_darwin_amd64
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-linux-musl-gcc go build -o bin/lrcAPI_v$1_linux_amd64
GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-musl-gcc go build -o bin/lrcAPI_v$1_linux_arm64
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o bin/lrcAPI_v$1_windows_amd64.exe
GOOS=windows GOARCH=arm64 CGO_ENABLED=1 CC=/opt/llvm-mingw/bin/aarch64-w64-mingw32-gcc go build -o bin/lrcAPI_v$1_windows_arm64.exe



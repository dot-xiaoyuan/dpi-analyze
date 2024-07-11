#!/bin/bash

set -e

OS=$(uname -s)
# TODO 根据操作系统构建打包脚本
echo "$OS"
#if [ "$OS" = "Darwin" ]; then
#    # MacOS
#    echo "Building for macOS..."
#    go build -o dpi-analyze ./cmd/dpi/main
#fi

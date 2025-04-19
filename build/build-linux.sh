#!/usr/bin/env bash
set -euo pipefail

#######################################
# build_linux.sh
#
# Usage: ./build_linux.sh <version>
# Example: ./build_linux.sh v1.2.3
#######################################

#######################################
# 0. Argument parsing
#######################################
if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <version>"
  exit 1
fi
VERSION="$1"

#######################################
# Variables (edit as needed)
#######################################
APP_NAME="M45-Relay-Client"

#######################################
# 1. Build for Linux (amd64), embedding flags
#######################################
# - publicClientFlag=true
# - version set to $VERSION
rm -f "${APP_NAME}"
GOOS=linux GOARCH=amd64 go build \
  -ldflags "\
    -X main.publicClientFlag=true \
    -X main.version=${VERSION}" \
  -o "${APP_NAME}"

#######################################
# 2. Zip the binary + readmes
#######################################
rm -f "${APP_NAME}-Linux.zip"
zip "${APP_NAME}-Linux.zip" "${APP_NAME}" readme.txt READ-ME.html

# Cleanup
rm -f "${APP_NAME}"

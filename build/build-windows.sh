#!/usr/bin/env bash
set -euo pipefail

#######################################
# build_windows.sh
#
# Usage: ./build_windows.sh <version>
# Example: ./build_windows.sh v2.0.0
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
# 1. Build for Windows (amd64),
#    embedding flags
#######################################
rm -f "${APP_NAME}.exe"
GOOS=windows GOARCH=amd64 go build \
  -ldflags "\
    -X main.publicClientFlag=true \
    -X main.version=${VERSION}" \
  -o "${APP_NAME}.exe"

#######################################
# 2. Zip the .exe + readmes
#######################################
ZIP_NAME="${APP_NAME}-Win.zip"
rm -f "${ZIP_NAME}"
zip "${ZIP_NAME}" "${APP_NAME}.exe" readme.txt READ-ME.html

# Cleanup
rm -f "${APP_NAME}.exe"

echo "Built Windows binary version ${VERSION} → ${ZIP_NAME}"

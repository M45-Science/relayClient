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
PFX_FILE="selfsigned.pfx"
PFX_PASS="changeit"
CERT_CN="M45 Relay Client (Self‑Signed)"
TIMESTAMP_URL=""  # e.g. "http://timestamp.digicert.com" if you want a timestamp

#######################################
# 1. Build for Windows (amd64),
#    embedding ldflags
#######################################
UNSIGNED_EXE="${APP_NAME}.exe"
rm -f "${UNSIGNED_EXE}"
GOOS=windows GOARCH=amd64 go build \
  -ldflags "\
    -X main.publicClientFlag=true \
    -X main.version=${VERSION}" \
  -o "${UNSIGNED_EXE}"

#######################################
# 2. Ensure self‑signed cert/PFX exists
#######################################
if ! command -v openssl >/dev/null 2>&1; then
  echo "Error: openssl not found in PATH" >&2
  exit 1
fi
if ! command -v osslsigncode >/dev/null 2>&1; then
  echo "Error: osslsigncode not found in PATH" >&2
  exit 1
fi

if [[ ! -f "${PFX_FILE}" ]]; then
  echo "==> Generating self‑signed certificate and PFX"
  # create key + cert
  openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -subj "/CN=${CERT_CN}" \
    -keyout cert.key \
    -out cert.pem

  # bundle into PFX
  openssl pkcs12 -export \
    -inkey cert.key \
    -in cert.pem \
    -out "${PFX_FILE}" \
    -passout "pass:${PFX_PASS}"

  # cleanup intermediate
  rm -f cert.key cert.pem
fi

#######################################
# 3. Sign the Windows binary
#######################################
SIGNED_EXE="${APP_NAME}-signed.exe"
echo "==> Signing ${UNSIGNED_EXE} → ${SIGNED_EXE}"
osslsigncode sign \
  -pkcs12 "${PFX_FILE}" \
  -pass "${PFX_PASS}" \
  -n "${CERT_CN}" \
  ${TIMESTAMP_URL:+-t "${TIMESTAMP_URL}"} \
  -in "${UNSIGNED_EXE}" \
  -out "${SIGNED_EXE}"

# replace unsigned with signed for packaging
rm -f "${UNSIGNED_EXE}"
mv "${SIGNED_EXE}" "${UNSIGNED_EXE}"

#######################################
# 4. Zip the signed .exe + readmes
#######################################
ZIP_NAME="${APP_NAME}-Win.zip"
rm -f "${ZIP_NAME}"
zip "${ZIP_NAME}" "${UNSIGNED_EXE}" readme.txt READ-ME.html

# Cleanup exe after zipping
rm -f "${UNSIGNED_EXE}"

echo "✔ Built & signed Windows version ${VERSION} → ${ZIP_NAME}"

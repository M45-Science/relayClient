#!/usr/bin/env bash
set -eu pipefail

#######################################
# wrapper.sh
#
# Usage:
#   # from project root
#   ./wrapper.sh v1.2.3
#
#   # or from inside build/
#   ../build/wrapper.sh v1.2.3
#######################################

# 1. Locate the directory this script resides in
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 2. Figure out where the build‑scripts folder is
#    If SCRIPT_DIR is itself "build", use it.
#    Otherwise assume a "build" subdirectory.
if [ "$(basename "${SCRIPT_DIR}")" = "build" ]; then
  BUILD_DIR="${SCRIPT_DIR}"
else
  BUILD_DIR="${SCRIPT_DIR}/build"
fi

# 3. Compute the project root = parent of BUILD_DIR
PROJECT_ROOT="$(dirname "${BUILD_DIR}")"

# 4. Parse your one required argument
if [ "$#" -ne 1 ]; then
  echo "Usage: $(basename "$0") <build‑arg>"
  exit 1
fi
BUILD_ARG="$1"

# 5. Invoke each script **from** PROJECT_ROOT, pointing at the scripts in BUILD_DIR
(
  cd "${PROJECT_ROOT}"
  bash "${BUILD_DIR}/build-linux.sh"  "${BUILD_ARG}"
)
(
  cd "${PROJECT_ROOT}"
  bash "${BUILD_DIR}/build-mac.sh"    "${BUILD_ARG}"
)
(
  cd "${PROJECT_ROOT}"
  bash "${BUILD_DIR}/build-windows.sh" "${BUILD_ARG}"
)

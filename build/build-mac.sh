#!/usr/bin/env bash

# Exit immediately on errors
set -e

#######################################
# Variables (edit as needed)
#######################################
APP_NAME="M45-Relay-Client"
APP_VERSION="1.0"
BINARY_NAME="M45-Relay-Client"    # The final universal (fat) binary name
WRAPPER_SCRIPT="run.sh"           # Shell script that launches Terminal
IDENTIFIER="com.M45.M45-Relay-Client"
MAIN_GO_FILE="."

#######################################
# 1. Build for macOS (amd64 & arm64),
#    then create a universal binary.
#######################################
echo "Building for macOS amd64..."
env GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.PublicClientMode=true" -o "${BINARY_NAME}-amd64" "${MAIN_GO_FILE}"

echo "Building for macOS arm64..."
env GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.PublicClientMode=true" -o "${BINARY_NAME}-arm64" "${MAIN_GO_FILE}"

echo "Combining into universal (fat) binary with lipo..."
llvm-lipo-14 -create "${BINARY_NAME}-amd64" "${BINARY_NAME}-arm64" -output "${BINARY_NAME}"

# Clean up the intermediate binaries
rm -f "${BINARY_NAME}-amd64" "${BINARY_NAME}-arm64"

#######################################
# 2. Create .app structure
#######################################
rm -rf "${APP_NAME}.app"  # Clean up any old build
mkdir -p "${APP_NAME}.app/Contents/MacOS"
mkdir -p "${APP_NAME}.app/Contents/Resources"

#######################################
# 3. Move the universal binary into the bundle
#######################################
mv "${BINARY_NAME}" "${APP_NAME}.app/Contents/MacOS/${BINARY_NAME}"

#######################################
# 4. Create a wrapper script that opens
#    Terminal and executes the binary
#######################################
cat <<EOF > "${APP_NAME}.app/Contents/MacOS/${WRAPPER_SCRIPT}"
#!/bin/bash
# Resolve the directory of this script so we know where the Go binary is
SELF_DIR="\$(cd "\$(dirname "\${BASH_SOURCE[0]}")" && pwd)"

# Use AppleScript to tell Terminal to run the Go binary in a new window
osascript <<OSASCRIPT
tell application "Terminal"
    activate
    do script "\\\"\$SELF_DIR/${BINARY_NAME}\\\""
end tell
OSASCRIPT

# Exit the wrapper script
exit 0
EOF

chmod +x "${APP_NAME}.app/Contents/MacOS/${WRAPPER_SCRIPT}"

#######################################
# 5. Create Info.plist referencing
#    the wrapper script as the executable
#######################################
cat <<EOF > "${APP_NAME}.app/Contents/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" 
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <!-- Application Name -->
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>

    <!-- Display Name -->
    <key>CFBundleDisplayName</key>
    <string>${APP_NAME}</string>

    <!-- The executable inside Contents/MacOS -->
    <key>CFBundleExecutable</key>
    <string>${WRAPPER_SCRIPT}</string>

    <!-- Bundle Identifier -->
    <key>CFBundleIdentifier</key>
    <string>${IDENTIFIER}</string>

    <!-- Version -->
    <key>CFBundleVersion</key>
    <string>${APP_VERSION}</string>

    <!-- Type (APPL = application) -->
    <key>CFBundlePackageType</key>
    <string>APPL</string>

    <key>CFBundleSignature</key>
    <string>????</string>

    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
  </dict>
</plist>
EOF

#######################################
# 6. Zip the .app + readmes
#######################################
echo "Zipping into M45-Relay-Mac.zip..."
rm -f ${APP_NAME}-Mac.zip
zip -r "${APP_NAME}-Mac.zip" "${APP_NAME}.app" readme.txt READ-ME.html
rm -rf "${APP_NAME}.app"

#######################################
# Done
#######################################
echo "Successfully built universal binary and created ${APP_NAME}.app"
echo "Output archive: M45-Relay-Mac.zip"

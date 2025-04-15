#!/bin/bash
# build_linux.sh

# Exit immediately on errors
set -e

#######################################
# Variables (edit as needed)
#######################################
APP_NAME="M45-Relay-Client"

#######################################
# 1. Build for Linux (amd64)
#######################################
rm -f ${APP_NAME}
GOOS=linux GOARCH=amd64 go build -ldflags="-X main.PublicClientMode=true" -o ${APP_NAME}
#######################################
# 2. Zip the .exe + readmes
#######################################
rm -f ${APP_NAME}-Linux.zip
zip ${APP_NAME}-Linux.zip ${APP_NAME} readme.txt READ-ME.html
rm -f ${APP_NAME}

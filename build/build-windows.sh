#!/bin/bash

# Exit immediately on errors
set -e

#######################################
# Variables (edit as needed)
#######################################
APP_NAME="M45-Relay-Client"

#######################################
# 1. Build for Windows (amd64)
#######################################
rm -f $APP_NAME
GOOS=windows GOARCH=amd64 go build -ldflags="-X main.PublicClientMode=true" -o ${APP_NAME}.exe

#######################################
# 2. Zip the .exe + readmes
#######################################
zip ${APP_NAME}-Win.zip ${APP_NAME}.exe readme.txt READ-ME.html readme.txt
rm -f ${APP_NAME}.exe
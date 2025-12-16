#!/bin/bash
# Downloads the onebusaway-gtfs-merge-cli JAR from Maven Central
# This JAR is used to compare Go implementation output against the original Java tool

set -e

VERSION="11.2.0"
JAR_NAME="onebusaway-gtfs-merge-cli-${VERSION}.jar"
JAR_PATH="$(dirname "$0")/${JAR_NAME}"
SYMLINK_PATH="$(dirname "$0")/onebusaway-gtfs-merge-cli.jar"

# Maven Central URL
URL="https://repo1.maven.org/maven2/org/onebusaway/onebusaway-gtfs-merge-cli/${VERSION}/${JAR_NAME}"

# Check if JAR already exists
if [ -f "$JAR_PATH" ]; then
    echo "JAR already exists: $JAR_PATH"
    exit 0
fi

echo "Downloading onebusaway-gtfs-merge-cli v${VERSION}..."
curl -fSL -o "$JAR_PATH" "$URL"

# Create symlink for version-agnostic access
ln -sf "$JAR_NAME" "$SYMLINK_PATH"

echo "Downloaded: $JAR_PATH"
echo "Symlinked: $SYMLINK_PATH"

# Verify the JAR works
if command -v java &> /dev/null; then
    echo "Verifying JAR..."
    java -jar "$JAR_PATH" --help > /dev/null 2>&1 && echo "JAR verified successfully" || echo "Warning: JAR verification failed (may need Java 11+)"
else
    echo "Warning: Java not found, skipping verification"
fi

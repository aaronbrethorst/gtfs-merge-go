#!/bin/bash
# Downloads real-world GTFS feeds for integration testing
# These feeds are from transit agencies in the Puget Sound region

set -e

SCRIPT_DIR="$(dirname "$0")"

# Parse arguments
FORCE=false
if [ "$1" = "--force" ]; then
    FORCE=true
    echo "Force mode: will re-download all feeds"
fi

# Feed definitions: name|url
FEEDS=(
    "pierce_transit|https://www.soundtransit.org/GTFS-PT/gtfs.zip"
    "intercity_transit|https://gtfs.sound.obaweb.org/prod/19_gtfs.zip"
    "community_transit|https://www.soundtransit.org/GTFS-CT/current.zip"
    "everett_transit|https://gtfs.sound.obaweb.org/prod/97_gtfs.zip"
)

download_count=0
skip_count=0

for feed in "${FEEDS[@]}"; do
    IFS='|' read -r name url <<< "$feed"
    path="$SCRIPT_DIR/${name}.zip"

    if [ -f "$path" ] && [ "$FORCE" = "false" ]; then
        echo "Already exists: $name.zip"
        ((skip_count++))
        continue
    fi

    echo "Downloading $name from $url..."
    if curl -fSL -o "$path" "$url"; then
        echo "  Downloaded: $name.zip ($(du -h "$path" | cut -f1))"
        ((download_count++))
    else
        echo "  ERROR: Failed to download $name"
        rm -f "$path"
        exit 1
    fi
done

echo ""
echo "Done. Downloaded: $download_count, Skipped: $skip_count"

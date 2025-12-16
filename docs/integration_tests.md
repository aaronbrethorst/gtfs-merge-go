# Real-World GTFS Feed Integration Tests

## Overview

Add integration tests using real transit agency GTFS feeds to validate that Java CLI and Go CLI produce **100% identical** CSV output for all merge scenarios.

## GTFS Feeds

| Agency | URL |
|--------|-----|
| Pierce Transit | https://www.soundtransit.org/GTFS-PT/gtfs.zip |
| Intercity Transit | https://gtfs.sound.obaweb.org/prod/19_gtfs.zip |
| Community Transit | https://www.soundtransit.org/GTFS-CT/current.zip |
| Everett Transit | https://gtfs.sound.obaweb.org/prod/97_gtfs.zip |

## Test Scenarios

12 scenarios total (selected pairs + multi-feed × 3 detection modes):

| Scenario | Feeds | Detection Modes |
|----------|-------|-----------------|
| 2-feed: Pierce + Intercity | 2 | none, identity, fuzzy |
| 2-feed: Community + Everett | 2 | none, identity, fuzzy |
| 3-feed merge | Pierce + Intercity + Community | none, identity, fuzzy |
| 4-feed merge | All 4 feeds | none, identity, fuzzy |

---

## Implementation Steps

### Step 1: Create Download Script

**File:** `testdata/realworld/download.sh`

```bash
#!/bin/bash
# Downloads real-world GTFS feeds for integration testing
set -e

SCRIPT_DIR="$(dirname "$0")"

FEEDS=(
    "pierce_transit|https://www.soundtransit.org/GTFS-PT/gtfs.zip"
    "intercity_transit|https://gtfs.sound.obaweb.org/prod/19_gtfs.zip"
    "community_transit|https://www.soundtransit.org/GTFS-CT/current.zip"
    "everett_transit|https://gtfs.sound.obaweb.org/prod/97_gtfs.zip"
)

for feed in "${FEEDS[@]}"; do
    IFS='|' read -r name url <<< "$feed"
    path="$SCRIPT_DIR/${name}.zip"

    if [ -f "$path" ]; then
        echo "Already exists: $path"
        continue
    fi

    echo "Downloading $name..."
    curl -fSL -o "$path" "$url"
done

echo "All feeds downloaded."
```

---

### Step 2: Add Diff-Style Output Formatter

**File:** `compare/compare.go` (add new function)

Add `FormatDiffStyleOutput()` function that formats differences in unified diff style:

```
--- expected/stops.txt
+++ actual/stops.txt
@@ 3 difference(s) @@
- [STOP_123] STOP_123,Main St Station,47.6,...
+ [STOP_123] STOP_123,Main Street Station,47.6,...
! [STOP_456]
-   STOP_456,Downtown,...
+   STOP_456,Downtown Transit,...
```

---

### Step 3: Create Real-World Test File

**File:** `compare/realworld_test.go` (new file)

```go
//go:build java

package compare

// Table-driven test scenarios
var realWorldScenarios = []struct {
    name      string
    feeds     []string
    detection string
}{
    // 2-feed combinations
    {"pierce_intercity_none", []string{"pierce_transit", "intercity_transit"}, "none"},
    {"pierce_intercity_identity", []string{"pierce_transit", "intercity_transit"}, "identity"},
    {"pierce_intercity_fuzzy", []string{"pierce_transit", "intercity_transit"}, "fuzzy"},

    {"community_everett_none", []string{"community_transit", "everett_transit"}, "none"},
    {"community_everett_identity", []string{"community_transit", "everett_transit"}, "identity"},
    {"community_everett_fuzzy", []string{"community_transit", "everett_transit"}, "fuzzy"},

    // 3-feed merge
    {"three_feed_none", []string{"pierce_transit", "intercity_transit", "community_transit"}, "none"},
    {"three_feed_identity", []string{"pierce_transit", "intercity_transit", "community_transit"}, "identity"},
    {"three_feed_fuzzy", []string{"pierce_transit", "intercity_transit", "community_transit"}, "fuzzy"},

    // 4-feed merge (all)
    {"four_feed_none", []string{"pierce_transit", "intercity_transit", "community_transit", "everett_transit"}, "none"},
    {"four_feed_identity", []string{"pierce_transit", "intercity_transit", "community_transit", "everett_transit"}, "identity"},
    {"four_feed_fuzzy", []string{"pierce_transit", "intercity_transit", "community_transit", "everett_transit"}, "fuzzy"},
}

func TestRealWorld_JavaGoComparison(t *testing.T) {
    jarPath := skipIfNoJava(t)
    feedPaths := skipIfNoRealWorldFeeds(t)

    for _, scenario := range realWorldScenarios {
        t.Run(scenario.name, func(t *testing.T) {
            // Run Java merge
            // Run Go merge
            // Compare with CompareGTFS()
            // On failure: output FormatDiffStyleOutput() and t.Fatal()
        })
    }
}
```

**Helper functions:**
- `skipIfNoRealWorldFeeds(t)` - Skip if feeds not downloaded, return paths map
- `getRealWorldFeedDir()` - Get absolute path to `testdata/realworld/`

---

### Step 4: Update CI Workflow

**File:** `.github/workflows/ci.yml`

Add to `compare-java` job:

```yaml
compare-java:
  runs-on: ubuntu-latest
  timeout-minutes: 60  # Increase for real-world feeds
  steps:
    # ... existing steps ...

    - name: Cache Real-World GTFS Feeds
      uses: actions/cache@v4
      with:
        path: testdata/realworld/*.zip
        key: realworld-gtfs-feeds-v1

    - name: Download Real-World GTFS Feeds
      run: ./testdata/realworld/download.sh

    - name: Run Java comparison tests
      run: go test -v -tags=java -timeout=60m ./compare/...
```

---

## Files to Create/Modify

| File | Action |
|------|--------|
| `testdata/realworld/download.sh` | Create - download script |
| `compare/compare.go` | Modify - add `FormatDiffStyleOutput()` |
| `compare/realworld_test.go` | Create - test file with scenarios |
| `.github/workflows/ci.yml` | Modify - add feed caching and download |

---

## Test Failure Output Example

When Java and Go outputs differ, the test will fail with:

```
=== RUN   TestRealWorld_JavaGoComparison/pierce_intercity_identity
    realworld_test.go:85: Java and Go outputs differ:
        --- expected/stops.txt
        +++ actual/stops.txt
        @@ 2 difference(s) @@
        - [a_STOP_001] a_STOP_001,Pierce Station,...
        + [a_STOP_001] a_STOP_001,Pierce Transit Station,...

--- FAIL: TestRealWorld_JavaGoComparison/pierce_intercity_identity (45.23s)
```

---

## Resource Considerations (Implementation Details)

### 1. Java CLI Memory

**File:** `compare/realworld_test.go`

The `JavaMerger` struct already has `MaxMemory` field (default "512m"). For real-world tests, set to "1g":

```go
javaMerger := NewJavaMerger(jarPath)
javaMerger.MaxMemory = "1g"  // Increase from default 512m for large feeds
```

### 2. CI Timeout

**File:** `.github/workflows/ci.yml`

Add timeout to `compare-java` job:

```yaml
compare-java:
  runs-on: ubuntu-latest
  timeout-minutes: 60  # Default is 360, but explicit is clearer
```

And to the test run command:

```yaml
- name: Run Java comparison tests
  run: go test -v -tags=java -timeout=60m ./compare/...
```

### 3. Feed Caching in CI

**File:** `.github/workflows/ci.yml`

Cache feeds between CI runs to avoid re-downloading:

```yaml
- name: Cache Real-World GTFS Feeds
  uses: actions/cache@v4
  with:
    path: testdata/realworld/*.zip
    key: realworld-gtfs-feeds-v1
    restore-keys: |
      realworld-gtfs-feeds-
```

The `v1` suffix allows cache invalidation by bumping to `v2` if feeds need refresh.

### 4. Parallel Test Execution

**File:** `compare/realworld_test.go`

Run scenarios in parallel for faster execution:

```go
func TestRealWorld_JavaGoComparison(t *testing.T) {
    jarPath := skipIfNoJava(t)
    feedPaths := skipIfNoRealWorldFeeds(t)

    for _, scenario := range realWorldScenarios {
        scenario := scenario // capture range variable
        t.Run(scenario.name, func(t *testing.T) {
            t.Parallel()  // Run each scenario concurrently
            // ... test logic
        })
    }
}
```

### 5. Temp Directory Cleanup

**File:** `compare/realworld_test.go`

Use `t.TempDir()` for automatic cleanup:

```go
tmpDir := t.TempDir()  // Automatically cleaned up after test
javaOutput := filepath.Join(tmpDir, "java_merged.zip")
goOutput := filepath.Join(tmpDir, "go_merged.zip")
```

### 6. Download Script Idempotency

**File:** `testdata/realworld/download.sh`

Script checks if files exist before downloading:

```bash
if [ -f "$path" ]; then
    echo "Already exists: $path"
    continue
fi
```

Add `--force` flag for manual refresh:

```bash
FORCE=false
if [ "$1" = "--force" ]; then
    FORCE=true
fi

# In download loop:
if [ -f "$path" ] && [ "$FORCE" = "false" ]; then
    echo "Already exists: $path"
    continue
fi
```

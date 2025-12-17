# Real-World GTFS Feed Integration Tests

## Overview

Integration tests using real transit agency GTFS feeds to validate that Java CLI and Go CLI produce comparable CSV output for merge scenarios.

## Status

| Component | Status |
|-----------|--------|
| Download script | Done |
| Test infrastructure | Done |
| CI configuration | Done |
| feed_info.txt parity | Partial (see Known Differences) |
| shapes.txt parity | Done |
| Full Java parity | In Progress |

---

## GTFS Feeds

| Agency | URL |
|--------|-----|
| Pierce Transit | https://www.soundtransit.org/GTFS-PT/gtfs.zip |
| Intercity Transit | https://gtfs.sound.obaweb.org/prod/19_gtfs.zip |
| Community Transit | https://www.soundtransit.org/GTFS-CT/current.zip |
| Everett Transit | https://gtfs.sound.obaweb.org/prod/97_gtfs.zip |

---

## Test Scenarios

12 scenarios total (selected pairs + multi-feed x 3 detection modes):

| Scenario | Feeds | Detection Modes |
|----------|-------|-----------------|
| 2-feed: Pierce + Intercity | 2 | none, identity, fuzzy |
| 2-feed: Community + Everett | 2 | none, identity, fuzzy |
| 3-feed merge | Pierce + Intercity + Community | none, identity, fuzzy |
| 4-feed merge | All 4 feeds | none, identity, fuzzy |

---

## Implemented Files

| File | Description |
|------|-------------|
| `testdata/realworld/download.sh` | Downloads GTFS feeds from transit agencies |
| `compare/realworld_test.go` | Integration test with all scenarios |
| `compare/compare.go` | `FormatDiffStyleOutput()` for diff-style failure output |
| `.github/workflows/ci.yml` | CI with feed caching and 60m timeout |

---

## Fixes Made for Java Parity

### 1. ID Prefix Delimiter (commit in earlier session)
- **Issue**: Go used `a_` prefix, Java uses `a-`
- **Fix**: Changed prefix delimiter in `merge/context.go`

### 2. feed_info.txt (commit 4c19d9f)
- **Issue**: Missing `feed_id` field, hardcoded columns
- **Fix**:
  - Added `FeedID` field to `FeedInfo` struct
  - Updated parser to read `feed_id` column
  - Updated writer with dynamic column output
  - Changed column tracking from intersection to union

### 3. shapes.txt (commit 8030334)
- **Issue**: 465K+ row differences due to sequence numbering
- **Root cause**: Java replaces `shape_pt_sequence` with a global counter; Go preserved original sequences
- **Fix**:
  - Added `ShapeSequenceCounter` to `MergeContext`
  - Added `NextShapeSequence()` method
  - Updated `ShapeMergeStrategy` to use global counter

---

## Known Differences (Remaining)

The following differences remain between Java and Go output:

### 1. Column Selection Approach

| Aspect | Go | Java |
|--------|-----|------|
| Column handling | Union (all columns from any feed) | Selective (appears to filter some columns) |
| Empty values | Outputs `0` or empty for missing data | Omits columns entirely |

**Affected files:**
- `stop_times.txt` - Go outputs `pickup_type`, `drop_off_type` columns; Java doesn't
- `routes.txt` - Go outputs `route_sort_order` with `0`; Java outputs empty
- `fare_rules.txt` - Go outputs `origin_id`, `destination_id`, `contains_id`; Java doesn't

### 2. feed_info.txt Row Count

| Aspect | Go | Java |
|--------|-----|------|
| Rows output | 1 (last feed wins) | Multiple (one per input feed) |
| feed_id assignment | Preserves from source | Assigns if missing |

### 3. Float Precision

Minor differences in float formatting for `shape_dist_traveled` and coordinates.

---

## Running Tests Locally

```bash
# Download GTFS feeds (first time only)
./testdata/realworld/download.sh

# Run all Java comparison tests
export JAVA_HOME=/path/to/java/21
go test -v -tags=java -timeout=60m ./compare/...

# Run a single scenario
go test -v -tags=java ./compare/... -run "TestRealWorld_JavaGoComparison/pierce_intercity_none$"
```

---

## CI Configuration

The `compare-java` job in `.github/workflows/ci.yml`:
- Caches downloaded GTFS feeds between runs
- Sets 60-minute timeout for large feed processing
- Runs all comparison tests with `-tags=java`

```yaml
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

## Future Work

To achieve 100% Java parity, the following would need investigation:

1. **Column selection logic**: Determine exactly how Java decides which columns to output (may involve examining Java source for column filtering logic)

2. **Multiple feed_info rows**: Consider changing `FeedInfo` from single struct to slice to support multiple rows like Java

3. **Float formatting**: Standardize float precision between implementations

4. **Test normalization**: Enhance the comparison normalizer to handle more edge cases

---

## Test Failure Output Example

When Java and Go outputs differ, the test fails with diff-style output:

```
=== RUN   TestRealWorld_JavaGoComparison/pierce_intercity_none
    realworld_test.go:91: Java and Go outputs differ:
        --- expected/stop_times.txt
        +++ actual/stop_times.txt
        @@ 441589 difference(s) @@
        ! [header]
        -   trip_id,arrival_time,departure_time,stop_id,stop_sequence,stop_headsign,shape_dist_traveled,timepoint
        +   trip_id,arrival_time,departure_time,stop_id,stop_sequence,stop_headsign,pickup_type,drop_off_type,shape_dist_traveled,timepoint
        ...
--- FAIL: TestRealWorld_JavaGoComparison/pierce_intercity_none (45.23s)
```

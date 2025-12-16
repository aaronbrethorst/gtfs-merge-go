package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

func TestFeedInfoMerge(t *testing.T) {
	// Given: source has feed info, target is empty
	source := gtfs.NewFeed()
	source.FeedInfo = &gtfs.FeedInfo{
		PublisherName: "Transit Authority",
		PublisherURL:  "http://transit.example.com",
		Lang:          "en",
		StartDate:     "20240101",
		EndDate:       "20241231",
		Version:       "1.0",
	}

	target := gtfs.NewFeed()

	ctx := NewMergeContext(source, target, "")
	strategy := NewFeedInfoMergeStrategy()

	// When: merged
	err := strategy.Merge(ctx)

	// Then: target should have feed info
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if target.FeedInfo == nil {
		t.Fatal("Expected feed info to be copied")
	}

	if target.FeedInfo.PublisherName != "Transit Authority" {
		t.Errorf("Expected PublisherName = Transit Authority, got %q", target.FeedInfo.PublisherName)
	}
}

func TestFeedInfoMergeCombinesVersions(t *testing.T) {
	// Given: both feeds have different versions
	source := gtfs.NewFeed()
	source.FeedInfo = &gtfs.FeedInfo{
		PublisherName: "Transit Authority",
		Version:       "2.0",
	}

	target := gtfs.NewFeed()
	target.FeedInfo = &gtfs.FeedInfo{
		PublisherName: "Transit Authority",
		Version:       "1.0",
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewFeedInfoMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged
	err := strategy.Merge(ctx)

	// Then: versions should be combined
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if target.FeedInfo.Version != "1.0, 2.0" {
		t.Errorf("Expected combined versions, got %q", target.FeedInfo.Version)
	}
}

func TestFeedInfoMergeExpandsDateRange(t *testing.T) {
	// Given: feeds have different date ranges
	source := gtfs.NewFeed()
	source.FeedInfo = &gtfs.FeedInfo{
		PublisherName: "Transit Authority",
		StartDate:     "20230601", // Earlier start
		EndDate:       "20241231", // Same end
	}

	target := gtfs.NewFeed()
	target.FeedInfo = &gtfs.FeedInfo{
		PublisherName: "Transit Authority",
		StartDate:     "20240101", // Later start
		EndDate:       "20250630", // Later end
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewFeedInfoMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged
	err := strategy.Merge(ctx)

	// Then: date range should be expanded to cover both
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if target.FeedInfo.StartDate != "20230601" {
		t.Errorf("Expected StartDate = 20230601 (earliest), got %q", target.FeedInfo.StartDate)
	}
	if target.FeedInfo.EndDate != "20250630" {
		t.Errorf("Expected EndDate = 20250630 (latest), got %q", target.FeedInfo.EndDate)
	}
}

func TestFeedInfoMergeNoSource(t *testing.T) {
	// Given: source has no feed info
	source := gtfs.NewFeed()

	target := gtfs.NewFeed()
	target.FeedInfo = &gtfs.FeedInfo{
		PublisherName: "Target Transit",
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewFeedInfoMergeStrategy()

	// When: merged
	err := strategy.Merge(ctx)

	// Then: target should keep its feed info unchanged
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if target.FeedInfo.PublisherName != "Target Transit" {
		t.Errorf("Expected target to keep its feed info")
	}
}

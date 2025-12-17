package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

func TestFeedInfoMerge(t *testing.T) {
	// Given: source has feed info, target is empty
	source := gtfs.NewFeed()
	source.FeedInfos["1"] = &gtfs.FeedInfo{
		FeedID:        "1",
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

	if len(target.FeedInfos) != 1 {
		t.Fatalf("Expected 1 feed info, got %d", len(target.FeedInfos))
	}

	fi := target.FeedInfos["1"]
	if fi == nil {
		t.Fatal("Expected feed info to be copied")
	}

	if fi.PublisherName != "Transit Authority" {
		t.Errorf("Expected PublisherName = Transit Authority, got %q", fi.PublisherName)
	}
}

func TestFeedInfoMergeMultipleFeedInfos(t *testing.T) {
	// Given: source has two feed infos with different feed_ids
	source := gtfs.NewFeed()
	source.FeedInfos["1"] = &gtfs.FeedInfo{
		FeedID:        "1",
		PublisherName: "Transit Authority A",
	}
	source.FeedInfos["99"] = &gtfs.FeedInfo{
		FeedID:        "99",
		PublisherName: "Transit Authority B",
	}

	target := gtfs.NewFeed()

	ctx := NewMergeContext(source, target, "")
	strategy := NewFeedInfoMergeStrategy()

	// When: merged
	err := strategy.Merge(ctx)

	// Then: target should have both feed infos
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.FeedInfos) != 2 {
		t.Fatalf("Expected 2 feed infos, got %d", len(target.FeedInfos))
	}

	if target.FeedInfos["1"].PublisherName != "Transit Authority A" {
		t.Errorf("Expected PublisherName A, got %q", target.FeedInfos["1"].PublisherName)
	}
	if target.FeedInfos["99"].PublisherName != "Transit Authority B" {
		t.Errorf("Expected PublisherName B, got %q", target.FeedInfos["99"].PublisherName)
	}
}

func TestFeedInfoMergeSameFeedIdOverwrites(t *testing.T) {
	// Given: source and target have same feed_id (source overwrites target)
	source := gtfs.NewFeed()
	source.FeedInfos["1"] = &gtfs.FeedInfo{
		FeedID:        "1",
		PublisherName: "Source Transit",
		Version:       "2.0",
	}

	target := gtfs.NewFeed()
	target.FeedInfos["1"] = &gtfs.FeedInfo{
		FeedID:        "1",
		PublisherName: "Target Transit",
		Version:       "1.0",
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewFeedInfoMergeStrategy()

	// When: merged
	err := strategy.Merge(ctx)

	// Then: source should overwrite target (last-read wins)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.FeedInfos) != 1 {
		t.Fatalf("Expected 1 feed info, got %d", len(target.FeedInfos))
	}

	fi := target.FeedInfos["1"]
	if fi.PublisherName != "Source Transit" {
		t.Errorf("Expected source to overwrite target, got PublisherName=%q", fi.PublisherName)
	}
	if fi.Version != "2.0" {
		t.Errorf("Expected source version, got %q", fi.Version)
	}
}

func TestFeedInfoMergeDifferentFeedIds(t *testing.T) {
	// Given: source and target have different feed_ids
	source := gtfs.NewFeed()
	source.FeedInfos["99"] = &gtfs.FeedInfo{
		FeedID:        "99",
		PublisherName: "Source Transit",
	}

	target := gtfs.NewFeed()
	target.FeedInfos["1"] = &gtfs.FeedInfo{
		FeedID:        "1",
		PublisherName: "Target Transit",
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewFeedInfoMergeStrategy()

	// When: merged
	err := strategy.Merge(ctx)

	// Then: both should be present
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.FeedInfos) != 2 {
		t.Fatalf("Expected 2 feed infos, got %d", len(target.FeedInfos))
	}

	if target.FeedInfos["1"].PublisherName != "Target Transit" {
		t.Errorf("Expected Target Transit, got %q", target.FeedInfos["1"].PublisherName)
	}
	if target.FeedInfos["99"].PublisherName != "Source Transit" {
		t.Errorf("Expected Source Transit, got %q", target.FeedInfos["99"].PublisherName)
	}
}

func TestFeedInfoMergeNoSource(t *testing.T) {
	// Given: source has no feed info
	source := gtfs.NewFeed()

	target := gtfs.NewFeed()
	target.FeedInfos["1"] = &gtfs.FeedInfo{
		FeedID:        "1",
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

	if len(target.FeedInfos) != 1 {
		t.Fatalf("Expected 1 feed info, got %d", len(target.FeedInfos))
	}

	if target.FeedInfos["1"].PublisherName != "Target Transit" {
		t.Errorf("Expected target to keep its feed info")
	}
}

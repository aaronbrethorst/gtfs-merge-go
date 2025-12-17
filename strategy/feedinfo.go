package strategy

// FeedInfoMergeStrategy handles merging of feed info between feeds
type FeedInfoMergeStrategy struct {
	BaseStrategy
}

// NewFeedInfoMergeStrategy creates a new FeedInfoMergeStrategy
func NewFeedInfoMergeStrategy() *FeedInfoMergeStrategy {
	return &FeedInfoMergeStrategy{
		BaseStrategy: NewBaseStrategy("feed_info"),
	}
}

// Merge performs the merge operation for feed info.
// FeedInfo entries are keyed by feed_id. When source and target have the same
// feed_id, the source entry overwrites the target (last-read wins), matching
// Java's behavior.
func (s *FeedInfoMergeStrategy) Merge(ctx *MergeContext) error {
	// Merge source FeedInfos into target (later overwrites earlier for same key)
	for id, fi := range ctx.Source.FeedInfos {
		ctx.Target.FeedInfos[id] = fi
	}
	return nil
}

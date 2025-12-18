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
	// Iterate in insertion order to match Java output
	for _, id := range ctx.Source.FeedInfoOrder {
		fi := ctx.Source.FeedInfos[id]
		// Track order only for new entries
		if _, exists := ctx.Target.FeedInfos[id]; !exists {
			ctx.Target.FeedInfoOrder = append(ctx.Target.FeedInfoOrder, id)
		}
		ctx.Target.FeedInfos[id] = fi
	}
	return nil
}

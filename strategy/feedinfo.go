package strategy

import (
	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

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

// Merge performs the merge operation for feed info
func (s *FeedInfoMergeStrategy) Merge(ctx *MergeContext) error {
	if ctx.Source.FeedInfo == nil {
		return nil
	}

	if ctx.Target.FeedInfo == nil {
		// No existing feed info - just copy
		ctx.Target.FeedInfo = &gtfs.FeedInfo{
			PublisherName: ctx.Source.FeedInfo.PublisherName,
			PublisherURL:  ctx.Source.FeedInfo.PublisherURL,
			Lang:          ctx.Source.FeedInfo.Lang,
			DefaultLang:   ctx.Source.FeedInfo.DefaultLang,
			StartDate:     ctx.Source.FeedInfo.StartDate,
			EndDate:       ctx.Source.FeedInfo.EndDate,
			Version:       ctx.Source.FeedInfo.Version,
			ContactEmail:  ctx.Source.FeedInfo.ContactEmail,
			ContactURL:    ctx.Source.FeedInfo.ContactURL,
			FeedID:        ctx.Source.FeedInfo.FeedID,
		}
	} else if s.DuplicateDetection == DetectionIdentity {
		// Merge: combine versions and expand date ranges
		existing := ctx.Target.FeedInfo
		source := ctx.Source.FeedInfo

		// Combine versions
		if source.Version != "" && existing.Version != source.Version {
			if existing.Version != "" {
				existing.Version = existing.Version + ", " + source.Version
			} else {
				existing.Version = source.Version
			}
		}

		// Expand date range - take earliest start date
		if source.StartDate != "" && (existing.StartDate == "" || source.StartDate < existing.StartDate) {
			existing.StartDate = source.StartDate
		}

		// Expand date range - take latest end date
		if source.EndDate != "" && (existing.EndDate == "" || source.EndDate > existing.EndDate) {
			existing.EndDate = source.EndDate
		}
	}

	return nil
}

package merge

import "github.com/aaronbrethorst/gtfs-merge-go/strategy"

// Option configures a Merger
type Option func(*Merger)

// WithDebug enables debug output
func WithDebug(debug bool) Option {
	return func(m *Merger) {
		m.debug = debug
	}
}

// WithDefaultDetection sets default duplicate detection for all strategies
func WithDefaultDetection(d strategy.DuplicateDetection) Option {
	return func(m *Merger) {
		m.SetDuplicateDetectionForAll(d)
	}
}

// WithDefaultLogging sets default duplicate logging for all strategies
func WithDefaultLogging(l strategy.DuplicateLogging) Option {
	return func(m *Merger) {
		m.agencyStrategy.SetDuplicateLogging(l)
		m.areaStrategy.SetDuplicateLogging(l)
		m.stopStrategy.SetDuplicateLogging(l)
		m.calendarStrategy.SetDuplicateLogging(l)
		m.calendarDateStrategy.SetDuplicateLogging(l)
		m.routeStrategy.SetDuplicateLogging(l)
		m.shapeStrategy.SetDuplicateLogging(l)
		m.tripStrategy.SetDuplicateLogging(l)
		m.stopTimeStrategy.SetDuplicateLogging(l)
		m.frequencyStrategy.SetDuplicateLogging(l)
		m.transferStrategy.SetDuplicateLogging(l)
		m.pathwayStrategy.SetDuplicateLogging(l)
		m.fareAttrStrategy.SetDuplicateLogging(l)
		m.fareRuleStrategy.SetDuplicateLogging(l)
		m.feedInfoStrategy.SetDuplicateLogging(l)
	}
}

// WithDefaultRenaming sets default renaming strategy for all strategies
func WithDefaultRenaming(r strategy.RenamingStrategy) Option {
	return func(m *Merger) {
		m.agencyStrategy.SetRenamingStrategy(r)
		m.areaStrategy.SetRenamingStrategy(r)
		m.stopStrategy.SetRenamingStrategy(r)
		m.calendarStrategy.SetRenamingStrategy(r)
		m.calendarDateStrategy.SetRenamingStrategy(r)
		m.routeStrategy.SetRenamingStrategy(r)
		m.shapeStrategy.SetRenamingStrategy(r)
		m.tripStrategy.SetRenamingStrategy(r)
		m.stopTimeStrategy.SetRenamingStrategy(r)
		m.frequencyStrategy.SetRenamingStrategy(r)
		m.transferStrategy.SetRenamingStrategy(r)
		m.pathwayStrategy.SetRenamingStrategy(r)
		m.fareAttrStrategy.SetRenamingStrategy(r)
		m.fareRuleStrategy.SetRenamingStrategy(r)
		m.feedInfoStrategy.SetRenamingStrategy(r)
	}
}

// WithFileDetection sets duplicate detection for a specific GTFS file.
// This matches the Java CLI behavior where --file and --duplicateDetection
// are paired by index to apply detection mode to specific entity types.
func WithFileDetection(filename string, d strategy.DuplicateDetection) Option {
	return func(m *Merger) {
		m.SetDuplicateDetectionForFile(filename, d)
	}
}

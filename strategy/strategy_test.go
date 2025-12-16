package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// MockStrategy is a test implementation of EntityMergeStrategy
type MockStrategy struct {
	name               string
	duplicateDetection DuplicateDetection
	duplicateLogging   DuplicateLogging
	renamingStrategy   RenamingStrategy
	mergeCallCount     int
	lastContext        *MergeContext
}

func (m *MockStrategy) Name() string {
	return m.name
}

func (m *MockStrategy) Merge(ctx *MergeContext) error {
	m.mergeCallCount++
	m.lastContext = ctx
	return nil
}

func (m *MockStrategy) SetDuplicateDetection(d DuplicateDetection) {
	m.duplicateDetection = d
}

func (m *MockStrategy) SetDuplicateLogging(l DuplicateLogging) {
	m.duplicateLogging = l
}

func (m *MockStrategy) SetRenamingStrategy(r RenamingStrategy) {
	m.renamingStrategy = r
}

func (m *MockStrategy) GetDuplicateDetection() DuplicateDetection {
	return m.duplicateDetection
}

func (m *MockStrategy) GetDuplicateLogging() DuplicateLogging {
	return m.duplicateLogging
}

func (m *MockStrategy) GetRenamingStrategy() RenamingStrategy {
	return m.renamingStrategy
}

// Verify MockStrategy implements EntityMergeStrategy
var _ EntityMergeStrategy = (*MockStrategy)(nil)

func TestStrategyInterfaceCompliance(t *testing.T) {
	// Create a mock strategy
	mock := &MockStrategy{name: "test-strategy"}

	// Test Name()
	if name := mock.Name(); name != "test-strategy" {
		t.Errorf("Name() = %q, want %q", name, "test-strategy")
	}

	// Test SetDuplicateDetection
	mock.SetDuplicateDetection(DetectionFuzzy)
	if mock.GetDuplicateDetection() != DetectionFuzzy {
		t.Errorf("GetDuplicateDetection() = %v, want %v", mock.GetDuplicateDetection(), DetectionFuzzy)
	}

	// Test SetDuplicateLogging
	mock.SetDuplicateLogging(LogWarning)
	if mock.GetDuplicateLogging() != LogWarning {
		t.Errorf("GetDuplicateLogging() = %v, want %v", mock.GetDuplicateLogging(), LogWarning)
	}

	// Test SetRenamingStrategy
	mock.SetRenamingStrategy(RenameAgency)
	if mock.GetRenamingStrategy() != RenameAgency {
		t.Errorf("GetRenamingStrategy() = %v, want %v", mock.GetRenamingStrategy(), RenameAgency)
	}

	// Test Merge
	source := gtfs.NewFeed()
	target := gtfs.NewFeed()
	ctx := NewMergeContext(source, target, "a_")

	if err := mock.Merge(ctx); err != nil {
		t.Errorf("Merge() returned error: %v", err)
	}
	if mock.mergeCallCount != 1 {
		t.Errorf("Merge() call count = %d, want 1", mock.mergeCallCount)
	}
	if mock.lastContext != ctx {
		t.Error("Merge() received different context than expected")
	}
}

func TestMergeContextCreation(t *testing.T) {
	source := gtfs.NewFeed()
	target := gtfs.NewFeed()

	// Add some test data to source
	source.Agencies[gtfs.AgencyID("A1")] = &gtfs.Agency{
		ID:   "A1",
		Name: "Test Agency",
	}

	ctx := NewMergeContext(source, target, "test_")

	// Verify context fields
	if ctx.Source != source {
		t.Error("MergeContext.Source != expected source")
	}
	if ctx.Target != target {
		t.Error("MergeContext.Target != expected target")
	}
	if ctx.Prefix != "test_" {
		t.Errorf("MergeContext.Prefix = %q, want %q", ctx.Prefix, "test_")
	}

	// Verify maps are initialized
	if ctx.EntityByRawID == nil {
		t.Error("MergeContext.EntityByRawID is nil")
	}
	if ctx.AgencyIDMapping == nil {
		t.Error("MergeContext.AgencyIDMapping is nil")
	}
	if ctx.StopIDMapping == nil {
		t.Error("MergeContext.StopIDMapping is nil")
	}
	if ctx.RouteIDMapping == nil {
		t.Error("MergeContext.RouteIDMapping is nil")
	}
	if ctx.TripIDMapping == nil {
		t.Error("MergeContext.TripIDMapping is nil")
	}
	if ctx.ServiceIDMapping == nil {
		t.Error("MergeContext.ServiceIDMapping is nil")
	}
	if ctx.ShapeIDMapping == nil {
		t.Error("MergeContext.ShapeIDMapping is nil")
	}
	if ctx.FareIDMapping == nil {
		t.Error("MergeContext.FareIDMapping is nil")
	}
	if ctx.AreaIDMapping == nil {
		t.Error("MergeContext.AreaIDMapping is nil")
	}
}

func TestMergeContextResolvedDetection(t *testing.T) {
	source := gtfs.NewFeed()
	target := gtfs.NewFeed()
	ctx := NewMergeContext(source, target, "")

	// Default should be DetectionNone
	if ctx.ResolvedDetection != DetectionNone {
		t.Errorf("Default ResolvedDetection = %v, want %v", ctx.ResolvedDetection, DetectionNone)
	}

	// Can be set
	ctx.ResolvedDetection = DetectionFuzzy
	if ctx.ResolvedDetection != DetectionFuzzy {
		t.Errorf("ResolvedDetection = %v, want %v", ctx.ResolvedDetection, DetectionFuzzy)
	}
}

func TestBaseStrategy(t *testing.T) {
	// Test that BaseStrategy provides default implementations
	base := &BaseStrategy{
		name: "base-test",
	}

	// Test Name
	if name := base.Name(); name != "base-test" {
		t.Errorf("Name() = %q, want %q", name, "base-test")
	}

	// Test default values
	if base.DuplicateDetection != DetectionNone {
		t.Errorf("Default DuplicateDetection = %v, want %v", base.DuplicateDetection, DetectionNone)
	}
	if base.DuplicateLogging != LogNone {
		t.Errorf("Default DuplicateLogging = %v, want %v", base.DuplicateLogging, LogNone)
	}
	if base.RenamingStrategy != RenameContext {
		t.Errorf("Default RenamingStrategy = %v, want %v", base.RenamingStrategy, RenameContext)
	}

	// Test setters
	base.SetDuplicateDetection(DetectionIdentity)
	if base.DuplicateDetection != DetectionIdentity {
		t.Errorf("DuplicateDetection = %v, want %v", base.DuplicateDetection, DetectionIdentity)
	}

	base.SetDuplicateLogging(LogError)
	if base.DuplicateLogging != LogError {
		t.Errorf("DuplicateLogging = %v, want %v", base.DuplicateLogging, LogError)
	}

	base.SetRenamingStrategy(RenameAgency)
	if base.RenamingStrategy != RenameAgency {
		t.Errorf("RenamingStrategy = %v, want %v", base.RenamingStrategy, RenameAgency)
	}
}

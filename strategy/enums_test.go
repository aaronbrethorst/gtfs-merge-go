package strategy

import "testing"

func TestDuplicateDetectionValues(t *testing.T) {
	// Verify enum has expected values
	tests := []struct {
		value    DuplicateDetection
		expected int
	}{
		{DetectionNone, 0},
		{DetectionIdentity, 1},
		{DetectionFuzzy, 2},
	}

	for _, tt := range tests {
		if int(tt.value) != tt.expected {
			t.Errorf("DuplicateDetection value = %d, want %d", tt.value, tt.expected)
		}
	}
}

func TestDuplicateLoggingValues(t *testing.T) {
	// Verify enum has expected values
	tests := []struct {
		value    DuplicateLogging
		expected int
	}{
		{LogNone, 0},
		{LogWarning, 1},
		{LogError, 2},
	}

	for _, tt := range tests {
		if int(tt.value) != tt.expected {
			t.Errorf("DuplicateLogging value = %d, want %d", tt.value, tt.expected)
		}
	}
}

func TestRenamingStrategyValues(t *testing.T) {
	// Verify enum has expected values
	tests := []struct {
		value    RenamingStrategy
		expected int
	}{
		{RenameContext, 0},
		{RenameAgency, 1},
	}

	for _, tt := range tests {
		if int(tt.value) != tt.expected {
			t.Errorf("RenamingStrategy value = %d, want %d", tt.value, tt.expected)
		}
	}
}

func TestDuplicateDetectionString(t *testing.T) {
	tests := []struct {
		value    DuplicateDetection
		expected string
	}{
		{DetectionNone, "none"},
		{DetectionIdentity, "identity"},
		{DetectionFuzzy, "fuzzy"},
	}

	for _, tt := range tests {
		if got := tt.value.String(); got != tt.expected {
			t.Errorf("DuplicateDetection.String() = %q, want %q", got, tt.expected)
		}
	}
}

func TestDuplicateLoggingString(t *testing.T) {
	tests := []struct {
		value    DuplicateLogging
		expected string
	}{
		{LogNone, "none"},
		{LogWarning, "warning"},
		{LogError, "error"},
	}

	for _, tt := range tests {
		if got := tt.value.String(); got != tt.expected {
			t.Errorf("DuplicateLogging.String() = %q, want %q", got, tt.expected)
		}
	}
}

func TestRenamingStrategyString(t *testing.T) {
	tests := []struct {
		value    RenamingStrategy
		expected string
	}{
		{RenameContext, "context"},
		{RenameAgency, "agency"},
	}

	for _, tt := range tests {
		if got := tt.value.String(); got != tt.expected {
			t.Errorf("RenamingStrategy.String() = %q, want %q", got, tt.expected)
		}
	}
}

func TestParseDuplicateDetection(t *testing.T) {
	tests := []struct {
		input    string
		expected DuplicateDetection
		hasError bool
	}{
		{"none", DetectionNone, false},
		{"identity", DetectionIdentity, false},
		{"fuzzy", DetectionFuzzy, false},
		{"NONE", DetectionNone, false},         // case insensitive
		{"Identity", DetectionIdentity, false}, // case insensitive
		{"invalid", DetectionNone, true},
		{"", DetectionNone, true},
	}

	for _, tt := range tests {
		got, err := ParseDuplicateDetection(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("ParseDuplicateDetection(%q) expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseDuplicateDetection(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("ParseDuplicateDetection(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		}
	}
}

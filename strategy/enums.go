// Package strategy defines merge strategies for different GTFS entity types.
package strategy

import (
	"fmt"
	"strings"
)

// DuplicateDetection specifies how duplicates are detected
type DuplicateDetection int

const (
	// DetectionNone - entities are never considered duplicates
	DetectionNone DuplicateDetection = iota

	// DetectionIdentity - entities with same ID are duplicates
	DetectionIdentity

	// DetectionFuzzy - entities with similar properties are duplicates
	DetectionFuzzy
)

// String returns the string representation of DuplicateDetection
func (d DuplicateDetection) String() string {
	switch d {
	case DetectionNone:
		return "none"
	case DetectionIdentity:
		return "identity"
	case DetectionFuzzy:
		return "fuzzy"
	default:
		return fmt.Sprintf("DuplicateDetection(%d)", d)
	}
}

// ParseDuplicateDetection parses a string into a DuplicateDetection value
func ParseDuplicateDetection(s string) (DuplicateDetection, error) {
	switch strings.ToLower(s) {
	case "none":
		return DetectionNone, nil
	case "identity":
		return DetectionIdentity, nil
	case "fuzzy":
		return DetectionFuzzy, nil
	default:
		return DetectionNone, fmt.Errorf("invalid duplicate detection mode: %q", s)
	}
}

// DuplicateLogging specifies how to handle detected duplicates
type DuplicateLogging int

const (
	// LogNone - no logging when duplicates are detected
	LogNone DuplicateLogging = iota

	// LogWarning - log a warning when duplicates are detected
	LogWarning

	// LogError - return an error when duplicates are detected
	LogError
)

// String returns the string representation of DuplicateLogging
func (l DuplicateLogging) String() string {
	switch l {
	case LogNone:
		return "none"
	case LogWarning:
		return "warning"
	case LogError:
		return "error"
	default:
		return fmt.Sprintf("DuplicateLogging(%d)", l)
	}
}

// RenamingStrategy specifies how duplicate IDs are renamed
type RenamingStrategy int

const (
	// RenameContext - use context prefix (a-, b-, c-, etc. for up to 26 feeds, or 00-, 01-, etc. for more)
	RenameContext RenamingStrategy = iota

	// RenameAgency - use agency-based naming
	RenameAgency
)

// String returns the string representation of RenamingStrategy
func (r RenamingStrategy) String() string {
	switch r {
	case RenameContext:
		return "context"
	case RenameAgency:
		return "agency"
	default:
		return fmt.Sprintf("RenamingStrategy(%d)", r)
	}
}

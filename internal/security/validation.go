package security

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// pathTraversalRegex attempts to detect ../ or ..\ patterns.
	pathTraversalRegex = regexp.MustCompile(`\.\.[\/\\]`)
)

// ValidateInput validates a string parameter to ensure it doesn't contain
// common injection vectors like path traversal.
func ValidateInput(input string) error {
	if pathTraversalRegex.MatchString(input) {
		return fmt.Errorf("security violation: path traversal detected in input")
	}

	// In a real system, we might add command injection or SQL injection checks here,
	// though they are often context-dependent.
	if strings.Contains(input, "\x00") {
		return fmt.Errorf("security violation: null byte detected in input")
	}

	return nil
}

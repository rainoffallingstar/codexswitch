package types

import (
	"fmt"
	"strings"
)

const DefaultReasoningEffort = "medium"

var allowedReasoningEfforts = map[string]struct{}{
	"none":    {},
	"minimal": {},
	"low":     {},
	"medium":  {},
	"high":    {},
	"xhigh":   {},
}

// NormalizeReasoningEffort trims and lowercases a reasoning effort value, validating allowed values.
func NormalizeReasoningEffort(value string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return DefaultReasoningEffort, nil
	}
	if _, ok := allowedReasoningEfforts[v]; !ok {
		return "", fmt.Errorf("invalid reasoning effort %q (allowed: none, minimal, low, medium, high, xhigh)", value)
	}
	return v, nil
}

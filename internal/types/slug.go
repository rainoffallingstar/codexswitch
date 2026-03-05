package types

import (
	"fmt"
	"regexp"
	"strings"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9_-]{1,64}$`)

// NormalizeSlug trims and lowercases a provider slug, validating safe cross-platform characters.
func NormalizeSlug(value string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(value))
	if strings.HasPrefix(v, ".") || !slugPattern.MatchString(v) {
		return "", fmt.Errorf("invalid provider slug %q (allowed pattern: [a-z0-9_-], length 1-64, cannot start with '.')", value)
	}
	return v, nil
}

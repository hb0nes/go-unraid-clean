package scan

import (
	"fmt"
	"regexp"
	"strings"
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func normalizeTitle(title string) string {
	lower := strings.ToLower(strings.TrimSpace(title))
	return strings.Trim(nonAlnum.ReplaceAllString(lower, " "), " ")
}

func normalizeTitleYear(title string, year int) string {
	norm := normalizeTitle(title)
	if year > 0 {
		return fmt.Sprintf("%s %d", norm, year)
	}
	return norm
}

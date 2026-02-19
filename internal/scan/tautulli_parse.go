package scan

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

type historyEntry struct {
	MediaType        string
	Title            string
	Year             int
	Guid             string
	ParentGuid       string
	GrandparentGuid  string
	GrandparentTitle string
	PercentComplete  int
	When             time.Time
	User             string
	WatchSeconds     int64
}

func parseHistoryEntry(raw map[string]any) (historyEntry, bool) {
	entry := historyEntry{}
	entry.MediaType = getString(raw, "media_type")
	entry.Title = getString(raw, "title", "full_title")
	entry.GrandparentTitle = getString(raw, "grandparent_title")
	entry.Guid = getString(raw, "guid")
	entry.ParentGuid = getString(raw, "parent_guid")
	entry.GrandparentGuid = getString(raw, "grandparent_guid")
	entry.Year = getInt(raw, "year")
	entry.PercentComplete = getInt(raw, "percent_complete", "percent")
	entry.User = getString(raw, "user", "username", "friendly_name")

	when, ok := getUnixTime(raw, "date", "stopped", "started", "last_viewed_at")
	if !ok {
		return entry, false
	}
	entry.When = when
	entry.WatchSeconds = getWatchSeconds(raw, entry.PercentComplete)
	return entry, true
}

func getString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			switch v := val.(type) {
			case string:
				if v != "" {
					return v
				}
			case json.Number:
				if v.String() != "" {
					return v.String()
				}
			}
		}
	}
	return ""
}

func getInt(m map[string]any, keys ...string) int {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			if out, ok := coerceInt(val); ok {
				return out
			}
		}
	}
	return 0
}

func coerceInt(val any) (int, bool) {
	switch v := val.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case json.Number:
		if i, err := strconv.Atoi(v.String()); err == nil {
			return i, true
		}
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
	}
	return 0, false
}

func getUnixTime(m map[string]any, keys ...string) (time.Time, bool) {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			if seconds, ok := coerceInt64(val); ok {
				if seconds > 0 {
					return time.Unix(seconds, 0).UTC(), true
				}
			}
		}
	}
	return time.Time{}, false
}

func coerceInt64(val any) (int64, bool) {
	switch v := val.(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	case json.Number:
		if i, err := strconv.ParseInt(v.String(), 10, 64); err == nil {
			return i, true
		}
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i, true
		}
	}
	return 0, false
}

func getWatchSeconds(m map[string]any, percent int) int64 {
	if seconds := getDurationSeconds(m, "watch_duration", "watched_duration", "play_duration"); seconds > 0 {
		return seconds
	}
	if duration := getDurationSeconds(m, "duration"); duration > 0 && percent > 0 {
		return int64(float64(duration) * float64(percent) / 100)
	}
	return getDurationSeconds(m, "view_offset")
}

func getDurationSeconds(m map[string]any, keys ...string) int64 {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			if seconds, ok := coerceInt64(val); ok {
				if seconds >= 100000 {
					return seconds / 1000
				}
				return seconds
			}
		}
	}
	return 0
}

func extractIDsFromGuid(guid string) (tmdbID int, tvdbID int, imdbID string) {
	if guid == "" {
		return 0, 0, ""
	}
	clean := strings.TrimSpace(guid)
	clean = strings.TrimPrefix(clean, "com.plexapp.agents.")

	parts := strings.SplitN(clean, "://", 2)
	if len(parts) != 2 {
		return 0, 0, ""
	}
	provider := parts[0]
	idPart := parts[1]
	idPart = strings.SplitN(idPart, "?", 2)[0]

	switch provider {
	case "themoviedb", "tmdb":
		if id, err := strconv.Atoi(idPart); err == nil {
			return id, 0, ""
		}
	case "tvdb":
		if id, err := strconv.Atoi(idPart); err == nil {
			return 0, id, ""
		}
	case "imdb":
		return 0, 0, idPart
	}
	return 0, 0, ""
}

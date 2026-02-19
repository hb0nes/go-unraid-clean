package scan

import "time"

type activityIndex struct {
	moviesByTMDB     map[int]time.Time
	moviesByIMDB     map[string]time.Time
	moviesByTitleKey map[string]time.Time
	seriesByTVDB     map[int]time.Time
	seriesByIMDB     map[string]time.Time
	seriesByTitleKey map[string]time.Time
}

func newActivityIndex() *activityIndex {
	return &activityIndex{
		moviesByTMDB:     map[int]time.Time{},
		moviesByIMDB:     map[string]time.Time{},
		moviesByTitleKey: map[string]time.Time{},
		seriesByTVDB:     map[int]time.Time{},
		seriesByIMDB:     map[string]time.Time{},
		seriesByTitleKey: map[string]time.Time{},
	}
}

func (a *activityIndex) recordMovie(tmdbID int, imdbID string, titleKey string, when time.Time) {
	if tmdbID > 0 {
		recordTime(a.moviesByTMDB, tmdbID, when)
	}
	if imdbID != "" {
		recordTime(a.moviesByIMDB, imdbID, when)
	}
	if titleKey != "" {
		recordTime(a.moviesByTitleKey, titleKey, when)
	}
}

func (a *activityIndex) recordSeries(tvdbID int, imdbID string, titleKey string, when time.Time) {
	if tvdbID > 0 {
		recordTime(a.seriesByTVDB, tvdbID, when)
	}
	if imdbID != "" {
		recordTime(a.seriesByIMDB, imdbID, when)
	}
	if titleKey != "" {
		recordTime(a.seriesByTitleKey, titleKey, when)
	}
}

func (a *activityIndex) movieLastActivity(tmdbID int, imdbID string, titleKey string) *time.Time {
	if tmdbID > 0 {
		if t, ok := a.moviesByTMDB[tmdbID]; ok {
			return &t
		}
	}
	if imdbID != "" {
		if t, ok := a.moviesByIMDB[imdbID]; ok {
			return &t
		}
	}
	if titleKey != "" {
		if t, ok := a.moviesByTitleKey[titleKey]; ok {
			return &t
		}
	}
	return nil
}

func (a *activityIndex) seriesLastActivity(tvdbID int, imdbID string, titleKey string) *time.Time {
	if tvdbID > 0 {
		if t, ok := a.seriesByTVDB[tvdbID]; ok {
			return &t
		}
	}
	if imdbID != "" {
		if t, ok := a.seriesByIMDB[imdbID]; ok {
			return &t
		}
	}
	if titleKey != "" {
		if t, ok := a.seriesByTitleKey[titleKey]; ok {
			return &t
		}
	}
	return nil
}

func recordTime[K comparable](m map[K]time.Time, key K, when time.Time) {
	if prev, ok := m[key]; ok {
		if when.After(prev) {
			m[key] = when
		}
		return
	}
	m[key] = when
}

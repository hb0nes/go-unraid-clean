package scan

import "time"

type activityWindow struct {
	First time.Time
	Last  time.Time
}

type activityIndex struct {
	moviesByTMDB     map[int]activityWindow
	moviesByIMDB     map[string]activityWindow
	moviesByTitleKey map[string]activityWindow
	seriesByTVDB     map[int]activityWindow
	seriesByIMDB     map[string]activityWindow
	seriesByTitleKey map[string]activityWindow
}

func newActivityIndex() *activityIndex {
	return &activityIndex{
		moviesByTMDB:     map[int]activityWindow{},
		moviesByIMDB:     map[string]activityWindow{},
		moviesByTitleKey: map[string]activityWindow{},
		seriesByTVDB:     map[int]activityWindow{},
		seriesByIMDB:     map[string]activityWindow{},
		seriesByTitleKey: map[string]activityWindow{},
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
		if w, ok := a.moviesByTMDB[tmdbID]; ok {
			return &w.Last
		}
	}
	if imdbID != "" {
		if w, ok := a.moviesByIMDB[imdbID]; ok {
			return &w.Last
		}
	}
	if titleKey != "" {
		if w, ok := a.moviesByTitleKey[titleKey]; ok {
			return &w.Last
		}
	}
	return nil
}

func (a *activityIndex) seriesLastActivity(tvdbID int, imdbID string, titleKey string) *time.Time {
	if tvdbID > 0 {
		if w, ok := a.seriesByTVDB[tvdbID]; ok {
			return &w.Last
		}
	}
	if imdbID != "" {
		if w, ok := a.seriesByIMDB[imdbID]; ok {
			return &w.Last
		}
	}
	if titleKey != "" {
		if w, ok := a.seriesByTitleKey[titleKey]; ok {
			return &w.Last
		}
	}
	return nil
}

func (a *activityIndex) movieFirstActivity(tmdbID int, imdbID string, titleKey string) *time.Time {
	if tmdbID > 0 {
		if w, ok := a.moviesByTMDB[tmdbID]; ok {
			return &w.First
		}
	}
	if imdbID != "" {
		if w, ok := a.moviesByIMDB[imdbID]; ok {
			return &w.First
		}
	}
	if titleKey != "" {
		if w, ok := a.moviesByTitleKey[titleKey]; ok {
			return &w.First
		}
	}
	return nil
}

func (a *activityIndex) seriesFirstActivity(tvdbID int, imdbID string, titleKey string) *time.Time {
	if tvdbID > 0 {
		if w, ok := a.seriesByTVDB[tvdbID]; ok {
			return &w.First
		}
	}
	if imdbID != "" {
		if w, ok := a.seriesByIMDB[imdbID]; ok {
			return &w.First
		}
	}
	if titleKey != "" {
		if w, ok := a.seriesByTitleKey[titleKey]; ok {
			return &w.First
		}
	}
	return nil
}

func recordTime[K comparable](m map[K]activityWindow, key K, when time.Time) {
	if prev, ok := m[key]; ok {
		first := prev.First
		last := prev.Last
		if when.Before(first) {
			first = when
		}
		if when.After(last) {
			last = when
		}
		m[key] = activityWindow{First: first, Last: last}
		return
	}
	m[key] = activityWindow{First: when, Last: when}
}

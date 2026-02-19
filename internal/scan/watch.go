package scan

import "sort"

type userWatch struct {
	User    string
	Seconds int64
}

type watchIndex struct {
	moviesByTMDB     map[int]map[string]int64
	moviesByIMDB     map[string]map[string]int64
	moviesByTitleKey map[string]map[string]int64
	seriesByTVDB     map[int]map[string]int64
	seriesByIMDB     map[string]map[string]int64
	seriesByTitleKey map[string]map[string]int64
}

func newWatchIndex() *watchIndex {
	return &watchIndex{
		moviesByTMDB:     map[int]map[string]int64{},
		moviesByIMDB:     map[string]map[string]int64{},
		moviesByTitleKey: map[string]map[string]int64{},
		seriesByTVDB:     map[int]map[string]int64{},
		seriesByIMDB:     map[string]map[string]int64{},
		seriesByTitleKey: map[string]map[string]int64{},
	}
}

func (w *watchIndex) recordMovie(tmdbID int, imdbID string, titleKey string, user string, seconds int64) {
	if tmdbID > 0 {
		recordUserTotals(w.moviesByTMDB, tmdbID, user, seconds)
	}
	if imdbID != "" {
		recordUserTotals(w.moviesByIMDB, imdbID, user, seconds)
	}
	if titleKey != "" {
		recordUserTotals(w.moviesByTitleKey, titleKey, user, seconds)
	}
}

func (w *watchIndex) recordSeries(tvdbID int, imdbID string, titleKey string, user string, seconds int64) {
	if tvdbID > 0 {
		recordUserTotals(w.seriesByTVDB, tvdbID, user, seconds)
	}
	if imdbID != "" {
		recordUserTotals(w.seriesByIMDB, imdbID, user, seconds)
	}
	if titleKey != "" {
		recordUserTotals(w.seriesByTitleKey, titleKey, user, seconds)
	}
}

func (w *watchIndex) movieTopUsers(tmdbID int, imdbID string, titleKey string, limit int) []userWatch {
	if tmdbID > 0 {
		if totals, ok := w.moviesByTMDB[tmdbID]; ok {
			return topUsers(totals, limit)
		}
	}
	if imdbID != "" {
		if totals, ok := w.moviesByIMDB[imdbID]; ok {
			return topUsers(totals, limit)
		}
	}
	if titleKey != "" {
		if totals, ok := w.moviesByTitleKey[titleKey]; ok {
			return topUsers(totals, limit)
		}
	}
	return nil
}

func (w *watchIndex) seriesTopUsers(tvdbID int, imdbID string, titleKey string, limit int) []userWatch {
	if tvdbID > 0 {
		if totals, ok := w.seriesByTVDB[tvdbID]; ok {
			return topUsers(totals, limit)
		}
	}
	if imdbID != "" {
		if totals, ok := w.seriesByIMDB[imdbID]; ok {
			return topUsers(totals, limit)
		}
	}
	if titleKey != "" {
		if totals, ok := w.seriesByTitleKey[titleKey]; ok {
			return topUsers(totals, limit)
		}
	}
	return nil
}

func recordUserTotals[K comparable](m map[K]map[string]int64, key K, user string, seconds int64) {
	if user == "" || seconds <= 0 {
		return
	}
	entry, ok := m[key]
	if !ok {
		entry = map[string]int64{}
		m[key] = entry
	}
	entry[user] += seconds
}

func topUsers(totals map[string]int64, limit int) []userWatch {
	if len(totals) == 0 {
		return nil
	}
	users := make([]userWatch, 0, len(totals))
	for user, seconds := range totals {
		users = append(users, userWatch{User: user, Seconds: seconds})
	}
	sort.Slice(users, func(i, j int) bool {
		if users[i].Seconds == users[j].Seconds {
			return users[i].User < users[j].User
		}
		return users[i].Seconds > users[j].Seconds
	})
	if limit <= 0 || len(users) <= limit {
		return users
	}
	return users[:limit]
}

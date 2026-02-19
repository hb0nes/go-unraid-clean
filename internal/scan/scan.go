package scan

import (
	"context"
	"fmt"
	"time"

	"go-unraid-clean/internal/clients"
	"go-unraid-clean/internal/config"
	"go-unraid-clean/internal/report"
)

const (
	reasonWatchInactive = "watch_inactive"
	reasonNeverWatched  = "never_watched"
)

func Run(ctx context.Context, cfg config.Config) (*report.Report, error) {
	radarr, err := clients.NewRadarrClient(cfg.Radarr.BaseURL, cfg.Radarr.APIKey)
	if err != nil {
		return nil, err
	}
	sonarr, err := clients.NewSonarrClient(cfg.Sonarr.BaseURL, cfg.Sonarr.APIKey)
	if err != nil {
		return nil, err
	}
	tautulli, err := clients.NewTautulliClient(cfg.Tautulli.BaseURL, cfg.Tautulli.APIKey)
	if err != nil {
		return nil, err
	}

	movies, err := radarr.Movies(ctx)
	if err != nil {
		return nil, err
	}
	series, err := sonarr.Series(ctx)
	if err != nil {
		return nil, err
	}
	entries, err := tautulli.History(ctx)
	if err != nil {
		return nil, err
	}

	activity := buildActivityIndex(entries, cfg.Rules.ActivityMinPercent)
	exceptions := newExceptionIndex(cfg)

	rep := &report.Report{
		GeneratedAt: time.Now().UTC(),
		Items:       []report.Item{},
	}
	cutoffWatch := time.Duration(cfg.Rules.InactivityDaysAfterWatch) * 24 * time.Hour
	cutoffNever := time.Duration(cfg.Rules.NeverWatchedDaysSinceAdded) * 24 * time.Hour

	now := time.Now().UTC()

	for _, movie := range movies {
		if !movie.HasFile || movie.SizeOnDisk == 0 {
			continue
		}
		if exceptions.isMovieException(movie.ID, movie.TMDBID, movie.IMDBID, movie.Title, movie.Path) {
			continue
		}

		lastActivity := activity.movieLastActivity(movie.TMDBID, movie.IMDBID, normalizeTitleYear(movie.Title, movie.Year))
		addedAt := parseTime(movie.Added)

		reason := evaluate(now, lastActivity, addedAt, cutoffWatch, cutoffNever)
		if reason == "" {
			continue
		}

		id := movie.ID
		rep.Items = append(rep.Items, report.Item{
			Type:           "movie",
			Title:          fmt.Sprintf("%s (%d)", movie.Title, movie.Year),
			RadarrID:       &id,
			Path:           movie.Path,
			SizeBytes:      movie.SizeOnDisk,
			AddedAt:        addedAt,
			LastActivityAt: lastActivity,
			Reason:         reason,
		})
	}

	for _, show := range series {
		if show.Statistics.SizeOnDisk == 0 {
			continue
		}
		if exceptions.isSeriesException(show.ID, show.TVDBID, show.IMDBID, show.Title, show.Path) {
			continue
		}

		lastActivity := activity.seriesLastActivity(show.TVDBID, show.IMDBID, normalizeTitle(show.Title))
		addedAt := parseTime(show.Added)
		reason := evaluate(now, lastActivity, addedAt, cutoffWatch, cutoffNever)
		if reason == "" {
			continue
		}

		id := show.ID
		rep.Items = append(rep.Items, report.Item{
			Type:           "series",
			Title:          show.Title,
			SonarrID:       &id,
			Path:           show.Path,
			SizeBytes:      show.Statistics.SizeOnDisk,
			AddedAt:        addedAt,
			LastActivityAt: lastActivity,
			Reason:         reason,
		})
	}

	return rep, nil
}

func buildActivityIndex(entries []map[string]any, minPercent int) *activityIndex {
	idx := newActivityIndex()
	for _, raw := range entries {
		entry, ok := parseHistoryEntry(raw)
		if !ok {
			continue
		}
		if entry.PercentComplete > 0 && entry.PercentComplete < minPercent {
			continue
		}

		switch entry.MediaType {
		case "movie":
			tmdbID, _, imdbID := extractIDsFromGuid(entry.Guid)
			titleKey := normalizeTitleYear(entry.Title, entry.Year)
			idx.recordMovie(tmdbID, imdbID, titleKey, entry.When)
		case "episode":
			_, tvdbID, imdbID := extractIDsFromGuid(entry.GrandparentGuid)
			if tvdbID == 0 {
				_, tvdbID, _ = extractIDsFromGuid(entry.ParentGuid)
			}
			if tvdbID == 0 {
				_, tvdbID, _ = extractIDsFromGuid(entry.Guid)
			}
			if imdbID == "" {
				_, _, imdbID = extractIDsFromGuid(entry.Guid)
			}
			titleKey := normalizeTitle(entry.GrandparentTitle)
			idx.recordSeries(tvdbID, imdbID, titleKey, entry.When)
		case "show":
			_, tvdbID, imdbID := extractIDsFromGuid(entry.Guid)
			titleKey := normalizeTitle(entry.Title)
			idx.recordSeries(tvdbID, imdbID, titleKey, entry.When)
		case "series":
			_, tvdbID, imdbID := extractIDsFromGuid(entry.Guid)
			titleKey := normalizeTitle(entry.Title)
			idx.recordSeries(tvdbID, imdbID, titleKey, entry.When)
		}
	}
	return idx
}

func evaluate(now time.Time, lastActivity *time.Time, addedAt *time.Time, cutoffWatch time.Duration, cutoffNever time.Duration) string {
	if lastActivity != nil {
		if now.Sub(*lastActivity) >= cutoffWatch {
			return reasonWatchInactive
		}
		return ""
	}
	if addedAt != nil && now.Sub(*addedAt) >= cutoffNever {
		return reasonNeverWatched
	}
	return ""
}

func parseTime(value string) *time.Time {
	if value == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		utc := t.UTC()
		return &utc
	}
	if t, err := time.Parse("2006-01-02T15:04:05", value); err == nil {
		utc := t.UTC()
		return &utc
	}
	return nil
}

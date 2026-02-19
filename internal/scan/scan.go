package scan

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go-unraid-clean/internal/clients"
	"go-unraid-clean/internal/config"
	"go-unraid-clean/internal/logging"
	"go-unraid-clean/internal/report"
)

const (
	reasonWatchInactive = "watch_inactive"
	reasonNeverWatched  = "never_watched"
)

type Options struct {
	SortBy    string
	SortOrder string
}

func Run(ctx context.Context, cfg config.Config, opts Options) (*report.Report, error) {
	log := logging.L()
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

	log.Info().Msg("Fetching Radarr movies")
	movies, err := radarr.Movies(ctx)
	if err != nil {
		return nil, err
	}
	log.Debug().Int("count", len(movies)).Msg("Loaded Radarr movies")
	log.Info().Msg("Fetching Sonarr series")
	series, err := sonarr.Series(ctx)
	if err != nil {
		return nil, err
	}
	log.Debug().Int("count", len(series)).Msg("Loaded Sonarr series")
	log.Info().Msg("Fetching Tautulli history")
	entries, err := tautulli.History(ctx)
	if err != nil {
		return nil, err
	}
	log.Debug().Int("count", len(entries)).Msg("Loaded Tautulli history entries")

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
			log.Debug().Str("title", movie.Title).Msg("Skipping movie due to exception")
			continue
		}

		titleKey := normalizeTitleYear(movie.Title, movie.Year)
		firstActivity := activity.movieFirstActivity(movie.TMDBID, movie.IMDBID, titleKey)
		lastActivity := activity.movieLastActivity(movie.TMDBID, movie.IMDBID, titleKey)
		addedAt := parseTime(movie.Added)

		reason := evaluate(now, lastActivity, addedAt, cutoffWatch, cutoffNever)
		if reason == "" {
			continue
		}

		id := movie.ID
		var tmdbPtr *int
		if movie.TMDBID > 0 {
			tmdb := movie.TMDBID
			tmdbPtr = &tmdb
		}
		rep.Items = append(rep.Items, report.Item{
			Type:            "movie",
			Title:           fmt.Sprintf("%s (%d)", movie.Title, movie.Year),
			RadarrID:        &id,
			TMDBID:          tmdbPtr,
			IMDBID:          movie.IMDBID,
			Path:            movie.Path,
			SizeBytes:       movie.SizeOnDisk,
			AddedAt:         addedAt,
			FirstActivityAt: firstActivity,
			LastActivityAt:  lastActivity,
			Reason:          reason,
		})
	}
	log.Info().Int("count", len(rep.Items)).Msg("Movies flagged for review")

	for _, show := range series {
		if show.Statistics.SizeOnDisk == 0 {
			continue
		}
		if exceptions.isSeriesException(show.ID, show.TVDBID, show.IMDBID, show.Title, show.Path) {
			log.Debug().Str("title", show.Title).Msg("Skipping series due to exception")
			continue
		}

		titleKey := normalizeTitle(show.Title)
		firstActivity := activity.seriesFirstActivity(show.TVDBID, show.IMDBID, titleKey)
		lastActivity := activity.seriesLastActivity(show.TVDBID, show.IMDBID, titleKey)
		addedAt := parseTime(show.Added)
		reason := evaluate(now, lastActivity, addedAt, cutoffWatch, cutoffNever)
		if reason == "" {
			continue
		}

		id := show.ID
		var tvdbPtr *int
		if show.TVDBID > 0 {
			tvdb := show.TVDBID
			tvdbPtr = &tvdb
		}
		rep.Items = append(rep.Items, report.Item{
			Type:            "series",
			Title:           show.Title,
			SonarrID:        &id,
			TVDBID:          tvdbPtr,
			IMDBID:          show.IMDBID,
			Path:            show.Path,
			SizeBytes:       show.Statistics.SizeOnDisk,
			AddedAt:         addedAt,
			FirstActivityAt: firstActivity,
			LastActivityAt:  lastActivity,
			Reason:          reason,
		})
	}
	log.Info().Int("count", len(rep.Items)).Msg("Total items flagged for review")

	if err := sortReport(rep, opts); err != nil {
		return nil, err
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

func sortReport(rep *report.Report, opts Options) error {
	sortBy := opts.SortBy
	if sortBy == "" {
		sortBy = "size"
	}
	order := opts.SortOrder
	if order == "" {
		order = "desc"
	}
	desc := order != "asc"

	switch sortBy {
	case "size":
		sort.SliceStable(rep.Items, func(i, j int) bool {
			if desc {
				return rep.Items[i].SizeBytes > rep.Items[j].SizeBytes
			}
			return rep.Items[i].SizeBytes < rep.Items[j].SizeBytes
		})
	case "added":
		sort.SliceStable(rep.Items, func(i, j int) bool {
			left := timeValue(rep.Items[i].AddedAt)
			right := timeValue(rep.Items[j].AddedAt)
			if desc {
				return left.After(right)
			}
			return left.Before(right)
		})
	case "gap":
		sort.SliceStable(rep.Items, func(i, j int) bool {
			left := gapDays(rep.Items[i].AddedAt, rep.Items[i].FirstActivityAt, rep.GeneratedAt)
			right := gapDays(rep.Items[j].AddedAt, rep.Items[j].FirstActivityAt, rep.GeneratedAt)
			if desc {
				return left > right
			}
			return left < right
		})
	case "last_activity":
		sort.SliceStable(rep.Items, func(i, j int) bool {
			left := timeValue(rep.Items[i].LastActivityAt)
			right := timeValue(rep.Items[j].LastActivityAt)
			if desc {
				return left.After(right)
			}
			return left.Before(right)
		})
	case "inactivity":
		sort.SliceStable(rep.Items, func(i, j int) bool {
			left := inactivityDays(rep.Items[i], rep.GeneratedAt)
			right := inactivityDays(rep.Items[j], rep.GeneratedAt)
			if desc {
				return left > right
			}
			return left < right
		})
	default:
		return fmt.Errorf("unsupported sort option: %s", sortBy)
	}
	return nil
}

func timeValue(val *time.Time) time.Time {
	if val == nil {
		return time.Time{}
	}
	return val.UTC()
}

func gapDays(addedAt *time.Time, firstActivityAt *time.Time, generatedAt time.Time) float64 {
	if addedAt == nil {
		return 0
	}
	end := generatedAt
	if firstActivityAt != nil {
		end = *firstActivityAt
	}
	span := end.Sub(*addedAt).Hours() / 24
	if span < 0 {
		span = 0
	}
	return span
}

func inactivityDays(item report.Item, generatedAt time.Time) float64 {
	if item.LastActivityAt != nil {
		span := generatedAt.Sub(*item.LastActivityAt).Hours() / 24
		if span < 0 {
			return 0
		}
		return span
	}
	if item.AddedAt != nil {
		span := generatedAt.Sub(*item.AddedAt).Hours() / 24
		if span < 0 {
			return 0
		}
		return span
	}
	return 0
}

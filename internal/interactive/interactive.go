package interactive

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go-unraid-clean/internal/clients"
	"go-unraid-clean/internal/config"
	"go-unraid-clean/internal/logging"
	"go-unraid-clean/internal/report"
)

func Run(ctx context.Context, cfgPath string, cfg config.Config, rep *report.Report) error {
	log := logging.L()
	radarr, err := clients.NewRadarrClient(cfg.Radarr.BaseURL, cfg.Radarr.APIKey)
	if err != nil {
		return err
	}
	sonarr, err := clients.NewSonarrClient(cfg.Sonarr.BaseURL, cfg.Sonarr.APIKey)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	changedConfig := false

	for idx, item := range rep.Items {
		fmt.Printf("\n[%d/%d] %s (%s)\n", idx+1, len(rep.Items), item.Title, item.Type)
		fmt.Printf("  Size: %s GiB\n", formatSizeGiB(item.SizeBytes))
		fmt.Printf("  Added: %s\n", formatOptionalTime(item.AddedAt))
		fmt.Printf("  First activity: %s\n", formatOptionalTime(item.FirstActivityAt))
		fmt.Printf("  Last activity: %s\n", formatOptionalTime(item.LastActivityAt))
		fmt.Printf("  Gap days: %s\n", formatGapDays(item.AddedAt, item.FirstActivityAt, rep.GeneratedAt))
		fmt.Printf("  Inactivity days: %s\n", formatInactivityDays(item.AddedAt, item.LastActivityAt, rep.GeneratedAt))
		fmt.Printf("  Reason: %s\n", item.Reason)
		fmt.Printf("  Path: %s\n", item.Path)

		options := "[k]eep [i]gnore [d]elete [f]delete-files [q]uit"
		if item.Type == "series" {
			options = "[k]eep [i]gnore [d]elete [f]delete-files [l]last-season [q]uit"
		}

		for {
			fmt.Printf("Action %s: ", options)
			input, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			choice := strings.ToLower(strings.TrimSpace(input))
			if choice == "" {
				continue
			}
			switch choice {
			case "k", "keep":
				log.Debug().Str("title", item.Title).Msg("Keeping item")
				goto nextItem
			case "i", "ignore", "safe":
				if err := addException(&cfg, item); err != nil {
					fmt.Printf("Failed to add exception: %s\n", err)
					continue
				}
				changedConfig = true
				fmt.Println("Added to exceptions.")
				goto nextItem
			case "d", "delete":
				if err := deleteItem(ctx, radarr, sonarr, item); err != nil {
					fmt.Printf("Delete failed: %s\n", err)
					continue
				}
				goto nextItem
			case "f", "files":
				if err := deleteFilesOnly(ctx, radarr, sonarr, item); err != nil {
					fmt.Printf("Delete files failed: %s\n", err)
					continue
				}
				goto nextItem
			case "l", "last":
				if item.Type != "series" {
					fmt.Println("Last-season is only valid for series.")
					continue
				}
				if err := keepLastSeason(ctx, sonarr, item); err != nil {
					fmt.Printf("Keep last season failed: %s\n", err)
					continue
				}
				goto nextItem
			case "q", "quit":
				if changedConfig {
					if err := config.Save(cfgPath, cfg); err != nil {
						return err
					}
					fmt.Printf("Saved config to %s\n", cfgPath)
				}
				return nil
			default:
				fmt.Println("Unknown action.")
			}
		}

	nextItem:
		continue
	}

	if changedConfig {
		if err := config.Save(cfgPath, cfg); err != nil {
			return err
		}
		fmt.Printf("Saved config to %s\n", cfgPath)
	}
	return nil
}

func addException(cfg *config.Config, item report.Item) error {
	switch item.Type {
	case "movie":
		if item.RadarrID != nil {
			cfg.Exceptions.Movies.RadarrIDs = config.AddUniqueInt(cfg.Exceptions.Movies.RadarrIDs, *item.RadarrID)
		}
		if item.TMDBID != nil {
			cfg.Exceptions.Movies.TMDBIDs = config.AddUniqueInt(cfg.Exceptions.Movies.TMDBIDs, *item.TMDBID)
		}
		if item.IMDBID != "" {
			cfg.Exceptions.Movies.IMDBIDs = config.AddUniqueString(cfg.Exceptions.Movies.IMDBIDs, item.IMDBID)
		}
		return nil
	case "series":
		if item.SonarrID != nil {
			cfg.Exceptions.Series.SonarrIDs = config.AddUniqueInt(cfg.Exceptions.Series.SonarrIDs, *item.SonarrID)
		}
		if item.TVDBID != nil {
			cfg.Exceptions.Series.TVDBIDs = config.AddUniqueInt(cfg.Exceptions.Series.TVDBIDs, *item.TVDBID)
		}
		if item.IMDBID != "" {
			cfg.Exceptions.Series.IMDBIDs = config.AddUniqueString(cfg.Exceptions.Series.IMDBIDs, item.IMDBID)
		}
		return nil
	default:
		return fmt.Errorf("unsupported item type: %s", item.Type)
	}
}

func deleteItem(ctx context.Context, radarr *clients.RadarrClient, sonarr *clients.SonarrClient, item report.Item) error {
	switch item.Type {
	case "movie":
		if item.RadarrID == nil {
			return fmt.Errorf("missing radarr_id")
		}
		return radarr.DeleteMovie(ctx, *item.RadarrID, true)
	case "series":
		if item.SonarrID == nil {
			return fmt.Errorf("missing sonarr_id")
		}
		return sonarr.DeleteSeries(ctx, *item.SonarrID, true)
	default:
		return fmt.Errorf("unsupported item type: %s", item.Type)
	}
}

func deleteFilesOnly(ctx context.Context, radarr *clients.RadarrClient, sonarr *clients.SonarrClient, item report.Item) error {
	switch item.Type {
	case "movie":
		if item.RadarrID == nil {
			return fmt.Errorf("missing radarr_id")
		}
		files, err := radarr.MovieFiles(ctx, *item.RadarrID)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			return fmt.Errorf("no movie files found")
		}
		for _, file := range files {
			if err := radarr.DeleteMovieFile(ctx, file.ID); err != nil {
				return err
			}
		}
		return nil
	case "series":
		if item.SonarrID == nil {
			return fmt.Errorf("missing sonarr_id")
		}
		ids, err := episodeFileIDs(ctx, sonarr, *item.SonarrID, 0)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return fmt.Errorf("no episode files found")
		}
		for _, id := range ids {
			if err := sonarr.DeleteEpisodeFile(ctx, id); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported item type: %s", item.Type)
	}
}

func keepLastSeason(ctx context.Context, sonarr *clients.SonarrClient, item report.Item) error {
	if item.SonarrID == nil {
		return fmt.Errorf("missing sonarr_id")
	}
	series, err := sonarr.SeriesByID(ctx, *item.SonarrID)
	if err != nil {
		return err
	}
	lastSeason := 0
	for _, season := range series.Seasons {
		if season.SeasonNumber == 0 {
			continue
		}
		if season.Statistics.EpisodeFileCount == 0 {
			continue
		}
		if season.SeasonNumber > lastSeason {
			lastSeason = season.SeasonNumber
		}
	}
	if lastSeason == 0 {
		return fmt.Errorf("no seasons with files found")
	}
	ids, err := episodeFileIDs(ctx, sonarr, *item.SonarrID, lastSeason)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return fmt.Errorf("no episode files found to delete")
	}
	for _, id := range ids {
		if err := sonarr.DeleteEpisodeFile(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func episodeFileIDs(ctx context.Context, sonarr *clients.SonarrClient, seriesID int, keepSeason int) ([]int, error) {
	episodes, err := sonarr.Episodes(ctx, seriesID)
	if err != nil {
		return nil, err
	}
	ids := map[int]struct{}{}
	for _, ep := range episodes {
		if ep.EpisodeFileID == 0 {
			continue
		}
		if ep.SeasonNumber == 0 {
			continue
		}
		if keepSeason > 0 && ep.SeasonNumber >= keepSeason {
			continue
		}
		ids[ep.EpisodeFileID] = struct{}{}
	}
	out := make([]int, 0, len(ids))
	for id := range ids {
		out = append(out, id)
	}
	return out, nil
}

func formatOptionalTime(val *time.Time) string {
	if val == nil {
		return ""
	}
	return val.UTC().Format(time.RFC3339)
}

func formatSizeGiB(bytes int64) string {
	if bytes <= 0 {
		return "0"
	}
	gib := float64(bytes) / (1024 * 1024 * 1024)
	return fmt.Sprintf("%.2f", gib)
}

func formatGapDays(addedAt *time.Time, firstActivityAt *time.Time, generatedAt time.Time) string {
	if addedAt == nil {
		return ""
	}
	end := generatedAt
	if firstActivityAt != nil {
		end = *firstActivityAt
	}
	span := end.Sub(*addedAt).Hours() / 24
	if span < 0 {
		span = 0
	}
	return fmt.Sprintf("%.1f", span)
}

func formatInactivityDays(addedAt *time.Time, lastActivityAt *time.Time, generatedAt time.Time) string {
	end := generatedAt
	if lastActivityAt != nil {
		end = *lastActivityAt
	} else if addedAt == nil {
		return ""
	}
	span := generatedAt.Sub(end).Hours() / 24
	if span < 0 {
		span = 0
	}
	return fmt.Sprintf("%.1f", span)
}

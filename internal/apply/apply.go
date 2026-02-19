package apply

import (
	"context"
	"errors"
	"fmt"

	"go-unraid-clean/internal/clients"
	"go-unraid-clean/internal/config"
	"go-unraid-clean/internal/logging"
	"go-unraid-clean/internal/report"
)

func Run(ctx context.Context, cfg config.Config, rep *report.Report) error {
	log := logging.L()
	radarr, err := clients.NewRadarrClient(cfg.Radarr.BaseURL, cfg.Radarr.APIKey)
	if err != nil {
		return err
	}
	sonarr, err := clients.NewSonarrClient(cfg.Sonarr.BaseURL, cfg.Sonarr.APIKey)
	if err != nil {
		return err
	}

	log.Info().Int("count", len(rep.Items)).Msg("Applying deletions")
	var errs []error
	for _, item := range rep.Items {
		switch item.Type {
		case "movie":
			if item.RadarrID == nil {
				errs = append(errs, fmt.Errorf("movie %q has no radarr_id", item.Title))
				continue
			}
			log.Info().Str("title", item.Title).Int("radarr_id", *item.RadarrID).Msg("Deleting movie")
			if err := radarr.DeleteMovie(ctx, *item.RadarrID, true); err != nil {
				errs = append(errs, err)
			}
		case "series":
			if item.SonarrID == nil {
				errs = append(errs, fmt.Errorf("series %q has no sonarr_id", item.Title))
				continue
			}
			log.Info().Str("title", item.Title).Int("sonarr_id", *item.SonarrID).Msg("Deleting series")
			if err := sonarr.DeleteSeries(ctx, *item.SonarrID, true); err != nil {
				errs = append(errs, err)
			}
		default:
			errs = append(errs, fmt.Errorf("unsupported item type %q for %s", item.Type, item.Title))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

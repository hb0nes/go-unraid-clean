package cmd

import (
	"context"
	"fmt"
	"strings"

	"go-unraid-clean/internal/clients"
	"go-unraid-clean/internal/config"

	"github.com/spf13/cobra"
)

var enrichDryRun bool

var enrichCmd = &cobra.Command{
	Use:   "enrich-exceptions",
	Short: "Backfill titles/paths for exception IDs",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}

		radarr, err := clients.NewRadarrClient(cfg.Radarr.BaseURL, cfg.Radarr.APIKey)
		if err != nil {
			return err
		}
		sonarr, err := clients.NewSonarrClient(cfg.Sonarr.BaseURL, cfg.Sonarr.APIKey)
		if err != nil {
			return err
		}

		movies, err := radarr.Movies(ctx)
		if err != nil {
			return err
		}
		series, err := sonarr.Series(ctx)
		if err != nil {
			return err
		}

		movieIndex := buildMovieIndex(movies)
		seriesIndex := buildSeriesIndex(series)

		changes := enrichExceptions(&cfg, movieIndex, seriesIndex)
		if len(changes) == 0 {
			fmt.Println("No exception entries to enrich.")
			return nil
		}

		for _, change := range changes {
			fmt.Println(change)
		}

		if enrichDryRun {
			fmt.Println("Dry run enabled; not writing config.")
			return nil
		}

		if err := config.Save(configPath, cfg); err != nil {
			return err
		}
		fmt.Printf("Updated config: %s\n", configPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(enrichCmd)
	enrichCmd.Flags().BoolVar(&enrichDryRun, "dry-run", false, "Show changes without writing config")
}

type movieIndex struct {
	byRadarr map[int]clients.RadarrMovie
	byTMDB   map[int]clients.RadarrMovie
	byIMDB   map[string]clients.RadarrMovie
}

type seriesIndex struct {
	bySonarr map[int]clients.SonarrSeries
	byTVDB   map[int]clients.SonarrSeries
	byIMDB   map[string]clients.SonarrSeries
}

func buildMovieIndex(movies []clients.RadarrMovie) movieIndex {
	idx := movieIndex{
		byRadarr: map[int]clients.RadarrMovie{},
		byTMDB:   map[int]clients.RadarrMovie{},
		byIMDB:   map[string]clients.RadarrMovie{},
	}
	for _, movie := range movies {
		idx.byRadarr[movie.ID] = movie
		if movie.TMDBID > 0 {
			idx.byTMDB[movie.TMDBID] = movie
		}
		if movie.IMDBID != "" {
			idx.byIMDB[strings.ToLower(movie.IMDBID)] = movie
		}
	}
	return idx
}

func buildSeriesIndex(series []clients.SonarrSeries) seriesIndex {
	idx := seriesIndex{
		bySonarr: map[int]clients.SonarrSeries{},
		byTVDB:   map[int]clients.SonarrSeries{},
		byIMDB:   map[string]clients.SonarrSeries{},
	}
	for _, show := range series {
		idx.bySonarr[show.ID] = show
		if show.TVDBID > 0 {
			idx.byTVDB[show.TVDBID] = show
		}
		if show.IMDBID != "" {
			idx.byIMDB[strings.ToLower(show.IMDBID)] = show
		}
	}
	return idx
}

func enrichExceptions(cfg *config.Config, movies movieIndex, series seriesIndex) []string {
	changes := []string{}

	for _, id := range cfg.Exceptions.Movies.RadarrIDs {
		if movie, ok := movies.byRadarr[id]; ok {
			if change := applyMovieDetails(cfg, movie); change != "" {
				changes = append(changes, change)
			}
		}
	}
	for _, id := range cfg.Exceptions.Movies.TMDBIDs {
		if movie, ok := movies.byTMDB[id]; ok {
			if change := applyMovieDetails(cfg, movie); change != "" {
				changes = append(changes, change)
			}
		}
	}
	for _, id := range cfg.Exceptions.Movies.IMDBIDs {
		if movie, ok := movies.byIMDB[strings.ToLower(id)]; ok {
			if change := applyMovieDetails(cfg, movie); change != "" {
				changes = append(changes, change)
			}
		}
	}

	for _, id := range cfg.Exceptions.Series.SonarrIDs {
		if show, ok := series.bySonarr[id]; ok {
			if change := applySeriesDetails(cfg, show); change != "" {
				changes = append(changes, change)
			}
		}
	}
	for _, id := range cfg.Exceptions.Series.TVDBIDs {
		if show, ok := series.byTVDB[id]; ok {
			if change := applySeriesDetails(cfg, show); change != "" {
				changes = append(changes, change)
			}
		}
	}
	for _, id := range cfg.Exceptions.Series.IMDBIDs {
		if show, ok := series.byIMDB[strings.ToLower(id)]; ok {
			if change := applySeriesDetails(cfg, show); change != "" {
				changes = append(changes, change)
			}
		}
	}

	return changes
}

func applyMovieDetails(cfg *config.Config, movie clients.RadarrMovie) string {
	changed := false
	before := len(cfg.Exceptions.Movies.Titles)
	cfg.Exceptions.Movies.Titles = config.AddUniqueString(cfg.Exceptions.Movies.Titles, movie.Title)
	changed = changed || len(cfg.Exceptions.Movies.Titles) > before

	before = len(cfg.Exceptions.Movies.PathPrefixes)
	cfg.Exceptions.Movies.PathPrefixes = config.AddUniqueString(cfg.Exceptions.Movies.PathPrefixes, movie.Path)
	changed = changed || len(cfg.Exceptions.Movies.PathPrefixes) > before

	before = len(cfg.Exceptions.Movies.IMDBIDs)
	cfg.Exceptions.Movies.IMDBIDs = config.AddUniqueString(cfg.Exceptions.Movies.IMDBIDs, movie.IMDBID)
	changed = changed || len(cfg.Exceptions.Movies.IMDBIDs) > before

	before = len(cfg.Exceptions.Movies.TMDBIDs)
	cfg.Exceptions.Movies.TMDBIDs = config.AddUniqueInt(cfg.Exceptions.Movies.TMDBIDs, movie.TMDBID)
	changed = changed || len(cfg.Exceptions.Movies.TMDBIDs) > before

	if !changed {
		return ""
	}
	return fmt.Sprintf("movie: %s", movie.Title)
}

func applySeriesDetails(cfg *config.Config, show clients.SonarrSeries) string {
	changed := false
	before := len(cfg.Exceptions.Series.Titles)
	cfg.Exceptions.Series.Titles = config.AddUniqueString(cfg.Exceptions.Series.Titles, show.Title)
	changed = changed || len(cfg.Exceptions.Series.Titles) > before

	before = len(cfg.Exceptions.Series.PathPrefixes)
	cfg.Exceptions.Series.PathPrefixes = config.AddUniqueString(cfg.Exceptions.Series.PathPrefixes, show.Path)
	changed = changed || len(cfg.Exceptions.Series.PathPrefixes) > before

	before = len(cfg.Exceptions.Series.IMDBIDs)
	cfg.Exceptions.Series.IMDBIDs = config.AddUniqueString(cfg.Exceptions.Series.IMDBIDs, show.IMDBID)
	changed = changed || len(cfg.Exceptions.Series.IMDBIDs) > before

	before = len(cfg.Exceptions.Series.TVDBIDs)
	cfg.Exceptions.Series.TVDBIDs = config.AddUniqueInt(cfg.Exceptions.Series.TVDBIDs, show.TVDBID)
	changed = changed || len(cfg.Exceptions.Series.TVDBIDs) > before

	if !changed {
		return ""
	}
	return fmt.Sprintf("series: %s", show.Title)
}

package config

import (
	"fmt"
	"net/url"
)

type Config struct {
	Tautulli   Service    `yaml:"tautulli"`
	Sonarr     Service    `yaml:"sonarr"`
	Radarr     Service    `yaml:"radarr"`
	Rules      Rules      `yaml:"rules"`
	Exceptions Exceptions `yaml:"exceptions"`
}

type Service struct {
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
}

type Rules struct {
	ActivityMinPercent         int     `yaml:"activity_min_percent"`
	InactivityDaysAfterWatch   int     `yaml:"inactivity_days_after_watch"`
	NeverWatchedDaysSinceAdded int     `yaml:"never_watched_days_since_added"`
	LowWatchMinAddedDays       int     `yaml:"low_watch_min_added_days"`
	LowWatchMaxHours           float64 `yaml:"low_watch_max_hours"`
	LowWatchRequire            bool    `yaml:"low_watch_require"`
	SeriesEndedOnly            bool    `yaml:"series_ended_only"`
}

type Exceptions struct {
	Movies MovieExceptions  `yaml:"movies"`
	Series SeriesExceptions `yaml:"series"`
}

type MovieExceptions struct {
	RadarrIDs    []int    `yaml:"radarr_ids"`
	TMDBIDs      []int    `yaml:"tmdb_ids"`
	IMDBIDs      []string `yaml:"imdb_ids"`
	Titles       []string `yaml:"titles"`
	PathPrefixes []string `yaml:"path_prefixes"`
}

type SeriesExceptions struct {
	SonarrIDs    []int    `yaml:"sonarr_ids"`
	TVDBIDs      []int    `yaml:"tvdb_ids"`
	IMDBIDs      []string `yaml:"imdb_ids"`
	Titles       []string `yaml:"titles"`
	PathPrefixes []string `yaml:"path_prefixes"`
}

func (c *Config) ApplyDefaults() {
	if c.Rules.ActivityMinPercent == 0 {
		c.Rules.ActivityMinPercent = 1
	}
	if c.Rules.InactivityDaysAfterWatch == 0 {
		c.Rules.InactivityDaysAfterWatch = 30
	}
	if c.Rules.NeverWatchedDaysSinceAdded == 0 {
		c.Rules.NeverWatchedDaysSinceAdded = 180
	}
}

func (c Config) Validate() error {
	if err := validateService("tautulli", c.Tautulli); err != nil {
		return err
	}
	if err := validateService("sonarr", c.Sonarr); err != nil {
		return err
	}
	if err := validateService("radarr", c.Radarr); err != nil {
		return err
	}
	if c.Rules.ActivityMinPercent <= 0 {
		return fmt.Errorf("rules: activity_min_percent must be positive")
	}
	if c.Rules.InactivityDaysAfterWatch <= 0 {
		return fmt.Errorf("rules: inactivity_days_after_watch must be positive")
	}
	if c.Rules.NeverWatchedDaysSinceAdded <= 0 {
		return fmt.Errorf("rules: never-watched days must be positive")
	}
	if c.Rules.LowWatchMinAddedDays < 0 || c.Rules.LowWatchMaxHours < 0 {
		return fmt.Errorf("rules: low watch thresholds must be non-negative")
	}
	if (c.Rules.LowWatchMinAddedDays > 0 && c.Rules.LowWatchMaxHours <= 0) ||
		(c.Rules.LowWatchMaxHours > 0 && c.Rules.LowWatchMinAddedDays <= 0) {
		return fmt.Errorf("rules: low_watch_min_added_days and low_watch_max_hours must both be set to enable")
	}
	return nil
}

func validateService(name string, svc Service) error {
	if svc.BaseURL == "" {
		return fmt.Errorf("%s: base_url is required", name)
	}
	if _, err := url.ParseRequestURI(svc.BaseURL); err != nil {
		return fmt.Errorf("%s: base_url is invalid: %w", name, err)
	}
	if svc.APIKey == "" {
		return fmt.Errorf("%s: api_key is required", name)
	}
	return nil
}

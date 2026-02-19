package scan

import (
	"path/filepath"
	"strings"

	"go-unraid-clean/internal/config"
)

type exceptionIndex struct {
	movieRadarrIDs map[int]struct{}
	movieTMDBIDs   map[int]struct{}
	movieIMDBIDs   map[string]struct{}
	movieTitles    map[string]struct{}
	moviePaths     []string

	seriesSonarrIDs map[int]struct{}
	seriesTVDBIDs   map[int]struct{}
	seriesIMDBIDs   map[string]struct{}
	seriesTitles    map[string]struct{}
	seriesPaths     []string
}

func newExceptionIndex(cfg config.Config) *exceptionIndex {
	idx := &exceptionIndex{
		movieRadarrIDs:  map[int]struct{}{},
		movieTMDBIDs:    map[int]struct{}{},
		movieIMDBIDs:    map[string]struct{}{},
		movieTitles:     map[string]struct{}{},
		moviePaths:      []string{},
		seriesSonarrIDs: map[int]struct{}{},
		seriesTVDBIDs:   map[int]struct{}{},
		seriesIMDBIDs:   map[string]struct{}{},
		seriesTitles:    map[string]struct{}{},
		seriesPaths:     []string{},
	}

	for _, id := range cfg.Exceptions.Movies.RadarrIDs {
		idx.movieRadarrIDs[id] = struct{}{}
	}
	for _, id := range cfg.Exceptions.Movies.TMDBIDs {
		idx.movieTMDBIDs[id] = struct{}{}
	}
	for _, id := range cfg.Exceptions.Movies.IMDBIDs {
		idx.movieIMDBIDs[strings.ToLower(id)] = struct{}{}
	}
	for _, title := range cfg.Exceptions.Movies.Titles {
		idx.movieTitles[normalizeTitle(title)] = struct{}{}
	}
	for _, prefix := range cfg.Exceptions.Movies.PathPrefixes {
		idx.moviePaths = append(idx.moviePaths, filepath.Clean(prefix))
	}

	for _, id := range cfg.Exceptions.Series.SonarrIDs {
		idx.seriesSonarrIDs[id] = struct{}{}
	}
	for _, id := range cfg.Exceptions.Series.TVDBIDs {
		idx.seriesTVDBIDs[id] = struct{}{}
	}
	for _, id := range cfg.Exceptions.Series.IMDBIDs {
		idx.seriesIMDBIDs[strings.ToLower(id)] = struct{}{}
	}
	for _, title := range cfg.Exceptions.Series.Titles {
		idx.seriesTitles[normalizeTitle(title)] = struct{}{}
	}
	for _, prefix := range cfg.Exceptions.Series.PathPrefixes {
		idx.seriesPaths = append(idx.seriesPaths, filepath.Clean(prefix))
	}

	return idx
}

func (e *exceptionIndex) isMovieException(radarrID int, tmdbID int, imdbID string, title string, path string) bool {
	if radarrID > 0 {
		if _, ok := e.movieRadarrIDs[radarrID]; ok {
			return true
		}
	}
	if tmdbID > 0 {
		if _, ok := e.movieTMDBIDs[tmdbID]; ok {
			return true
		}
	}
	if imdbID != "" {
		if _, ok := e.movieIMDBIDs[strings.ToLower(imdbID)]; ok {
			return true
		}
	}
	if title != "" {
		if _, ok := e.movieTitles[normalizeTitle(title)]; ok {
			return true
		}
	}
	return hasPathPrefix(path, e.moviePaths)
}

func (e *exceptionIndex) isSeriesException(sonarrID int, tvdbID int, imdbID string, title string, path string) bool {
	if sonarrID > 0 {
		if _, ok := e.seriesSonarrIDs[sonarrID]; ok {
			return true
		}
	}
	if tvdbID > 0 {
		if _, ok := e.seriesTVDBIDs[tvdbID]; ok {
			return true
		}
	}
	if imdbID != "" {
		if _, ok := e.seriesIMDBIDs[strings.ToLower(imdbID)]; ok {
			return true
		}
	}
	if title != "" {
		if _, ok := e.seriesTitles[normalizeTitle(title)]; ok {
			return true
		}
	}
	return hasPathPrefix(path, e.seriesPaths)
}

func hasPathPrefix(path string, prefixes []string) bool {
	if path == "" || len(prefixes) == 0 {
		return false
	}
	cleaned := filepath.Clean(path)
	for _, prefix := range prefixes {
		if strings.HasPrefix(cleaned, prefix) {
			return true
		}
	}
	return false
}

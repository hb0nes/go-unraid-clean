package clients

import (
	"context"
	"fmt"
	"net/http"
)

type SonarrClient struct {
	http *HTTPClient
}

type SonarrSeries struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Year       int    `json:"year"`
	TVDBID     int    `json:"tvdbId"`
	IMDBID     string `json:"imdbId"`
	Status     string `json:"status"`
	Path       string `json:"path"`
	Added      string `json:"added"`
	Statistics struct {
		SizeOnDisk int64 `json:"sizeOnDisk"`
	} `json:"statistics"`
}

type SonarrSeriesDetail struct {
	ID      int            `json:"id"`
	Title   string         `json:"title"`
	Seasons []SonarrSeason `json:"seasons"`
}

type SonarrSeason struct {
	SeasonNumber int `json:"seasonNumber"`
	Statistics   struct {
		EpisodeFileCount int `json:"episodeFileCount"`
	} `json:"statistics"`
}

type SonarrEpisode struct {
	ID            int `json:"id"`
	SeriesID      int `json:"seriesId"`
	SeasonNumber  int `json:"seasonNumber"`
	EpisodeFileID int `json:"episodeFileId"`
}

type SonarrEpisodeFile struct {
	ID           int `json:"id"`
	SeriesID     int `json:"seriesId"`
	SeasonNumber int `json:"seasonNumber"`
}

func NewSonarrClient(baseURL, apiKey string) (*SonarrClient, error) {
	hc, err := NewHTTPClient(baseURL, apiKey)
	if err != nil {
		return nil, err
	}
	return &SonarrClient{http: hc}, nil
}

func (c *SonarrClient) Series(ctx context.Context) ([]SonarrSeries, error) {
	url := c.http.Resolve("api/v3/series")
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", c.http.APIKey)

	resp, err := doRequest(ctx, c.http.Client, req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := readBody(resp)
		return nil, fmt.Errorf("sonarr series: status %d: %s", resp.StatusCode, string(body))
	}

	var out []SonarrSeries
	if err := decodeJSONBody(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *SonarrClient) SeriesByID(ctx context.Context, id int) (*SonarrSeriesDetail, error) {
	url := c.http.Resolve(fmt.Sprintf("api/v3/series/%d", id))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", c.http.APIKey)

	resp, err := doRequest(ctx, c.http.Client, req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := readBody(resp)
		return nil, fmt.Errorf("sonarr series %d: status %d: %s", id, resp.StatusCode, string(body))
	}

	var out SonarrSeriesDetail
	if err := decodeJSONBody(resp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SonarrClient) Episodes(ctx context.Context, seriesID int) ([]SonarrEpisode, error) {
	url := c.http.Resolve(fmt.Sprintf("api/v3/episode?seriesId=%d", seriesID))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", c.http.APIKey)

	resp, err := doRequest(ctx, c.http.Client, req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := readBody(resp)
		return nil, fmt.Errorf("sonarr episodes: status %d: %s", resp.StatusCode, string(body))
	}

	var out []SonarrEpisode
	if err := decodeJSONBody(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *SonarrClient) EpisodeFiles(ctx context.Context, seriesID int) ([]SonarrEpisodeFile, error) {
	url := c.http.Resolve(fmt.Sprintf("api/v3/episodefile?seriesId=%d", seriesID))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", c.http.APIKey)

	resp, err := doRequest(ctx, c.http.Client, req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := readBody(resp)
		return nil, fmt.Errorf("sonarr episode files: status %d: %s", resp.StatusCode, string(body))
	}

	var out []SonarrEpisodeFile
	if err := decodeJSONBody(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *SonarrClient) DeleteEpisodeFile(ctx context.Context, id int) error {
	url := c.http.Resolve(fmt.Sprintf("api/v3/episodefile/%d", id))
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", c.http.APIKey)

	resp, err := doRequest(ctx, c.http.Client, req)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := readBody(resp)
		return fmt.Errorf("sonarr delete episodefile %d: status %d: %s", id, resp.StatusCode, string(body))
	}
	return nil
}

func (c *SonarrClient) DeleteSeries(ctx context.Context, id int, deleteFiles bool) error {
	url := c.http.Resolve(fmt.Sprintf("api/v3/series/%d?deleteFiles=%t&addImportListExclusion=false", id, deleteFiles))
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", c.http.APIKey)

	resp, err := doRequest(ctx, c.http.Client, req)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := readBody(resp)
		return fmt.Errorf("sonarr delete series %d: status %d: %s", id, resp.StatusCode, string(body))
	}
	return nil
}

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
	Path       string `json:"path"`
	Added      string `json:"added"`
	Statistics struct {
		SizeOnDisk int64 `json:"sizeOnDisk"`
	} `json:"statistics"`
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

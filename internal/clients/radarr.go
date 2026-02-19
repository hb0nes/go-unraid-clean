package clients

import (
	"context"
	"fmt"
	"net/http"
)

type RadarrClient struct {
	http *HTTPClient
}

type RadarrMovie struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Year       int    `json:"year"`
	TMDBID     int    `json:"tmdbId"`
	IMDBID     string `json:"imdbId"`
	Path       string `json:"path"`
	Added      string `json:"added"`
	SizeOnDisk int64  `json:"sizeOnDisk"`
	HasFile    bool   `json:"hasFile"`
}

type RadarrMovieFile struct {
	ID           int    `json:"id"`
	MovieID      int    `json:"movieId"`
	Size         int64  `json:"size"`
	Quality      any    `json:"quality"`
	RelativePath string `json:"relativePath"`
}

func NewRadarrClient(baseURL, apiKey string) (*RadarrClient, error) {
	hc, err := NewHTTPClient(baseURL, apiKey)
	if err != nil {
		return nil, err
	}
	return &RadarrClient{http: hc}, nil
}

func (c *RadarrClient) Movies(ctx context.Context) ([]RadarrMovie, error) {
	url := c.http.Resolve("api/v3/movie")
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
		return nil, fmt.Errorf("radarr movies: status %d: %s", resp.StatusCode, string(body))
	}

	var out []RadarrMovie
	if err := decodeJSONBody(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *RadarrClient) MovieFiles(ctx context.Context, movieID int) ([]RadarrMovieFile, error) {
	url := c.http.Resolve(fmt.Sprintf("api/v3/moviefile?movieId=%d", movieID))
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
		return nil, fmt.Errorf("radarr movie files: status %d: %s", resp.StatusCode, string(body))
	}

	var out []RadarrMovieFile
	if err := decodeJSONBody(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *RadarrClient) DeleteMovieFile(ctx context.Context, id int) error {
	url := c.http.Resolve(fmt.Sprintf("api/v3/moviefile/%d", id))
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
		return fmt.Errorf("radarr delete movie file %d: status %d: %s", id, resp.StatusCode, string(body))
	}
	return nil
}

func (c *RadarrClient) DeleteMovie(ctx context.Context, id int, deleteFiles bool) error {
	url := c.http.Resolve(fmt.Sprintf("api/v3/movie/%d?deleteFiles=%t&addImportExclusion=false", id, deleteFiles))
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
		return fmt.Errorf("radarr delete movie %d: status %d: %s", id, resp.StatusCode, string(body))
	}
	return nil
}

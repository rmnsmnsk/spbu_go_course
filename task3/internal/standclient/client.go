package standclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"movie-service/internal/movie"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{},
		timeout:    timeout,
	}
}

func (c *Client) GetMovie(ctx context.Context, id int64) (movie.Movie, error) {
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	endpoint := fmt.Sprintf("%s/%d/info.0.json", c.baseURL, id)
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return movie.Movie{}, fmt.Errorf("create movie request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return movie.Movie{}, movie.ErrUpstreamTimeout
		}
		return movie.Movie{}, fmt.Errorf("request movie stand: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return movie.Movie{}, movie.ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return movie.Movie{}, fmt.Errorf("movie stand returned HTTP status %d", resp.StatusCode)
	}

	var result movie.Movie
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return movie.Movie{}, fmt.Errorf("decode movie %d: %w", id, err)
	}
	return result, nil
}

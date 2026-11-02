package moviestand

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"movie-service/internal/domain"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

type movieResponse struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Year     int32  `json:"year"`
	Director string `json:"director"`
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{},
		timeout:    timeout,
	}
}

func (c *Client) GetMovie(ctx context.Context, id int64) (domain.Movie, error) {
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	endpoint := fmt.Sprintf("%s/%d/info.0.json", c.baseURL, id)
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return domain.Movie{}, fmt.Errorf("create movie request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		switch {
		case errors.Is(requestCtx.Err(), context.DeadlineExceeded):
			return domain.Movie{}, domain.ErrUpstreamTimeout
		case errors.Is(requestCtx.Err(), context.Canceled):
			return domain.Movie{}, context.Canceled
		default:
			return domain.Movie{}, fmt.Errorf("request movie stand: %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return domain.Movie{}, domain.ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.Movie{}, fmt.Errorf("movie stand returned HTTP status %d", resp.StatusCode)
	}

	var response movieResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return domain.Movie{}, fmt.Errorf("decode movie %d: %w", id, err)
	}
	return domain.Movie{
		ID:       response.ID,
		Title:    response.Title,
		Year:     response.Year,
		Director: response.Director,
	}, nil
}

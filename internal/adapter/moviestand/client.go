package moviestand

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"movie-service/internal/domain"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	workers    int
}

type movieResponse struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Year     int32  `json:"year"`
	Director string `json:"director"`
}

func New(baseURL string, timeout time.Duration, workers int) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{},
		timeout:    timeout,
		workers:    workers,
	}
}

type movieResult struct {
	id    int64
	movie domain.Movie
	err   error
}

func (c *Client) GetMoviesRange(ctx context.Context, from, to int64) (domain.LoadResult, error) {
	if from <= 0 || to <= 0 || from > to {
		return domain.LoadResult{}, domain.ErrInvalidRange
	}

	ids := make([]int64, 0, to-from+1)
	for id := from; ; id++ {
		ids = append(ids, id)
		if id == to {
			break
		}
	}
	return c.GetMovies(ctx, ids)
}

func (c *Client) GetMovies(ctx context.Context, ids []int64) (domain.LoadResult, error) {
	result := domain.LoadResult{
		Movies: make([]domain.Movie, 0, len(ids)),
		Errors: make([]domain.LoadError, 0),
	}
	if len(ids) == 0 {
		return result, nil
	}

	jobs := make(chan int64)
	results := make(chan movieResult)
	workerCount := c.workers
	if len(ids) < workerCount {
		workerCount = len(ids)
	}

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case id, ok := <-jobs:
					if !ok {
						return
					}
					loaded, err := c.GetMovie(ctx, id)
					select {
					case results <- movieResult{id: id, movie: loaded, err: err}:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, id := range ids {
			select {
			case jobs <- id:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for loaded := range results {
		if loaded.err != nil {
			result.Errors = append(result.Errors, domain.LoadError{ID: loaded.id, Message: loaded.err.Error()})
			continue
		}
		result.Movies = append(result.Movies, loaded.movie)
	}

	if err := ctx.Err(); err != nil {
		return domain.LoadResult{}, err
	}
	sort.Slice(result.Movies, func(i, j int) bool { return result.Movies[i].ID < result.Movies[j].ID })
	sort.Slice(result.Errors, func(i, j int) bool { return result.Errors[i].ID < result.Errors[j].ID })
	return result, nil
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

func (c *Client) GetLastMovie(ctx context.Context) (domain.Movie, error) {
	first, err := c.GetMovie(ctx, 1)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.Movie{}, domain.ErrNoMovies
	}
	if err != nil {
		return domain.Movie{}, err
	}

	low, high, last, err := c.findSearchBounds(ctx, first)
	if err != nil {
		return domain.Movie{}, err
	}
	return c.findLastInRange(ctx, low, high, last)
}

func (c *Client) findSearchBounds(ctx context.Context, first domain.Movie) (int64, int64, domain.Movie, error) {
	low, high := int64(1), int64(2)
	last := first
	for {
		candidate, err := c.GetMovie(ctx, high)
		if errors.Is(err, domain.ErrNotFound) {
			return low, high, last, nil
		}
		if err != nil {
			return 0, 0, domain.Movie{}, err
		}
		low, last = high, candidate
		if high > math.MaxInt64/2 {
			return 0, 0, domain.Movie{}, errors.New("movie id exceeds supported range")
		}
		high *= 2
	}
}

func (c *Client) findLastInRange(ctx context.Context, low, high int64, last domain.Movie) (domain.Movie, error) {
	for low+1 < high {
		middle := low + (high-low)/2
		candidate, err := c.GetMovie(ctx, middle)
		switch {
		case err == nil:
			low, last = middle, candidate
		case errors.Is(err, domain.ErrNotFound):
			high = middle
		default:
			return domain.Movie{}, err
		}
	}
	return last, nil
}

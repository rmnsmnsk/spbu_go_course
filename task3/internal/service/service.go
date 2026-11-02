package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"

	"movie-service/internal/movie"
)

type MovieGetter interface {
	GetMovie(context.Context, int64) (movie.Movie, error)
}

type Service struct {
	client  MovieGetter
	workers int
}

func New(client MovieGetter, workers int) *Service {
	return &Service{client: client, workers: workers}
}

func ValidateRange(from, to int64) error {
	switch {
	case from <= 0 || to <= 0:
		return fmt.Errorf("%w: from and to must be positive", movie.ErrInvalidRange)
	case from > to:
		return fmt.Errorf("%w: from must not be greater than to", movie.ErrInvalidRange)
	default:
		return nil
	}
}

type jobResult struct {
	id    int64
	movie movie.Movie
	err   error
}

func (s *Service) LoadMovies(ctx context.Context, from, to int64) (movie.LoadResult, error) {
	if err := ValidateRange(from, to); err != nil {
		return movie.LoadResult{}, err
	}

	jobs := make(chan int64)
	results := make(chan jobResult)
	workerCount := s.workers
	if to-from < int64(workerCount) {
		workerCount = int(to - from + 1)
	}

	var wg sync.WaitGroup
	for range workerCount {
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
					loaded, err := s.client.GetMovie(ctx, id)
					select {
					case results <- jobResult{id: id, movie: loaded, err: err}:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for id := from; id <= to; id++ {
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

	response := movie.LoadResult{
		Movies: make([]movie.Movie, 0),
		Errors: make([]movie.LoadError, 0),
	}
	for result := range results {
		if result.err != nil {
			response.Errors = append(response.Errors, movie.LoadError{ID: result.id, Message: result.err.Error()})
			continue
		}
		response.Movies = append(response.Movies, result.movie)
	}

	if err := ctx.Err(); err != nil {
		return movie.LoadResult{}, err
	}
	sort.Slice(response.Movies, func(i, j int) bool { return response.Movies[i].ID < response.Movies[j].ID })
	sort.Slice(response.Errors, func(i, j int) bool { return response.Errors[i].ID < response.Errors[j].ID })
	return response, nil
}

func (s *Service) GetLastMovie(ctx context.Context) (movie.Movie, error) {
	first, err := s.client.GetMovie(ctx, 1)
	if errors.Is(err, movie.ErrNotFound) {
		return movie.Movie{}, movie.ErrNoMovies
	}
	if err != nil {
		return movie.Movie{}, err
	}

	low, high := int64(1), int64(2)
	last := first
	for {
		candidate, err := s.client.GetMovie(ctx, high)
		if errors.Is(err, movie.ErrNotFound) {
			break
		}
		if err != nil {
			return movie.Movie{}, err
		}
		low, last = high, candidate
		if high > math.MaxInt64/2 {
			return movie.Movie{}, errors.New("movie id exceeds supported range")
		}
		high *= 2
	}

	for low+1 < high {
		middle := low + (high-low)/2
		candidate, err := s.client.GetMovie(ctx, middle)
		switch {
		case err == nil:
			low, last = middle, candidate
		case errors.Is(err, movie.ErrNotFound):
			high = middle
		default:
			return movie.Movie{}, err
		}
	}
	return last, nil
}

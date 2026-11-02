package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"

	"movie-service/internal/domain"
)

type MovieProvider interface {
	GetMovie(ctx context.Context, id int64) (domain.Movie, error)
}

type MovieUseCase interface {
	LoadMovies(ctx context.Context, from, to int64) (domain.LoadResult, error)
	GetLastMovie(ctx context.Context) (domain.Movie, error)
}

type Service struct {
	provider MovieProvider
	workers  int
}

func New(provider MovieProvider, workers int) *Service {
	return &Service{provider: provider, workers: workers}
}

func validateRange(from, to int64) error {
	switch {
	case from <= 0 || to <= 0:
		return fmt.Errorf("%w: from and to must be positive", domain.ErrInvalidRange)
	case from > to:
		return fmt.Errorf("%w: from must not be greater than to", domain.ErrInvalidRange)
	default:
		return nil
	}
}

type jobResult struct {
	id    int64
	movie domain.Movie
	err   error
}

func (s *Service) LoadMovies(ctx context.Context, from, to int64) (domain.LoadResult, error) {
	if err := validateRange(from, to); err != nil {
		return domain.LoadResult{}, err
	}

	jobs := make(chan int64)
	results := make(chan jobResult)
	workerCount := s.workers
	if to-from < int64(workerCount) {
		workerCount = int(to - from + 1)
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
					loaded, err := s.provider.GetMovie(ctx, id)
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

	result := domain.LoadResult{
		Movies: make([]domain.Movie, 0),
		Errors: make([]domain.LoadError, 0),
	}
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

func (s *Service) GetLastMovie(ctx context.Context) (domain.Movie, error) {
	first, err := s.provider.GetMovie(ctx, 1)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.Movie{}, domain.ErrNoMovies
	}
	if err != nil {
		return domain.Movie{}, err
	}

	low, high := int64(1), int64(2)
	last := first
	for {
		candidate, err := s.provider.GetMovie(ctx, high)
		if errors.Is(err, domain.ErrNotFound) {
			break
		}
		if err != nil {
			return domain.Movie{}, err
		}
		low, last = high, candidate
		if high > math.MaxInt64/2 {
			return domain.Movie{}, errors.New("movie id exceeds supported range")
		}
		high *= 2
	}

	for low+1 < high {
		middle := low + (high-low)/2
		candidate, err := s.provider.GetMovie(ctx, middle)
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

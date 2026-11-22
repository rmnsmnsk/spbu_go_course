package usecase

import (
	"context"
	"fmt"

	"movie-service/internal/domain"
)

type MovieRepository interface {
	GetMoviesRange(ctx context.Context, from, to int64) (domain.LoadResult, error)
	GetMoviesByIDs(ctx context.Context, ids []int64) ([]domain.Movie, error)
	GetLastMovie(ctx context.Context) (domain.Movie, error)
}

type SearchIndex interface {
	FindAny(ctx context.Context, words []string) ([]int64, error)
}

type SearchCache interface {
	Get(ctx context.Context, query string) ([]domain.Movie, bool, error)
	Set(ctx context.Context, query string, movies []domain.Movie) error
}

type MovieUseCase interface {
	LoadMovies(ctx context.Context, from, to int64) (domain.LoadResult, error)
	GetLastMovie(ctx context.Context) (domain.Movie, error)
	SearchMovies(ctx context.Context, query string) ([]domain.Movie, error)
}

type Service struct {
	repository MovieRepository
	index      SearchIndex
	cache      SearchCache
}

func New(repository MovieRepository, index SearchIndex, cache SearchCache) *Service {
	return &Service{repository: repository, index: index, cache: cache}
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

func (s *Service) LoadMovies(ctx context.Context, from, to int64) (domain.LoadResult, error) {
	if err := validateRange(from, to); err != nil {
		return domain.LoadResult{}, err
	}
	return s.repository.GetMoviesRange(ctx, from, to)
}

func (s *Service) GetLastMovie(ctx context.Context) (domain.Movie, error) {
	return s.repository.GetLastMovie(ctx)
}

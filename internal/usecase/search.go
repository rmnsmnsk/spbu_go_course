package usecase

import (
	"context"
	"fmt"
	"strings"

	"movie-service/internal/domain"
)

func (s *Service) SearchMovies(ctx context.Context, query string) ([]domain.Movie, error) {
	words := splitWords(query)
	if len(words) == 0 {
		return nil, domain.ErrInvalidQuery
	}
	query = strings.ToLower(strings.TrimSpace(query))

	cachedMovies, found, err := s.cache.Get(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get cached search results: %w", err)
	}
	if found {
		return cachedMovies, nil
	}

	ids, err := s.index.FindAny(ctx, words)
	if err != nil {
		return nil, fmt.Errorf("search movie index: %w", err)
	}
	if len(ids) == 0 {
		movies := []domain.Movie{}
		if err := s.cache.Set(ctx, query, movies); err != nil {
			return nil, fmt.Errorf("cache search results: %w", err)
		}
		return movies, nil
	}

	movies, err := s.repository.GetMoviesByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("get search results: %w", err)
	}
	if err := s.cache.Set(ctx, query, movies); err != nil {
		return nil, fmt.Errorf("cache search results: %w", err)
	}
	return movies, nil
}

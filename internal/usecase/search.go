package usecase

import (
	"context"
	"fmt"

	"movie-service/internal/domain"
)

func (s *Service) SearchMovies(ctx context.Context, query string) ([]domain.Movie, error) {
	words := splitWords(query)
	if len(words) == 0 {
		return nil, domain.ErrInvalidQuery
	}

	ids, err := s.index.FindAny(ctx, words)
	if err != nil {
		return nil, fmt.Errorf("search movie index: %w", err)
	}
	if len(ids) == 0 {
		return []domain.Movie{}, nil
	}

	movies, err := s.repository.GetMoviesByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("get search results: %w", err)
	}
	return movies, nil
}

package usecase

import (
	"context"
	"errors"
	"fmt"

	"movie-service/internal/domain"
)

type MovieStand interface {
	GetLastMovie(ctx context.Context) (domain.Movie, error)
	GetMovies(ctx context.Context, ids []int64) (domain.LoadResult, error)
}

type MovieStorage interface {
	GetMovieIDs(ctx context.Context, to int64) ([]int64, error)
	SaveMovie(ctx context.Context, movie domain.Movie) error
}

type Synchronizer struct {
	stand   MovieStand
	storage MovieStorage
}

func NewSynchronizer(stand MovieStand, storage MovieStorage) *Synchronizer {
	return &Synchronizer{stand: stand, storage: storage}
}

func (s *Synchronizer) Sync(ctx context.Context) error {
	last, err := s.stand.GetLastMovie(ctx)
	if errors.Is(err, domain.ErrNoMovies) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("find last movie on stand: %w", err)
	}

	existingIDs, err := s.storage.GetMovieIDs(ctx, last.ID)
	if err != nil {
		return fmt.Errorf("get stored movie IDs: %w", err)
	}
	missingIDs := findMissingIDs(last.ID, existingIDs)
	if len(missingIDs) == 0 {
		return nil
	}

	result, err := s.stand.GetMovies(ctx, missingIDs)
	if err != nil {
		return fmt.Errorf("load missing movies: %w", err)
	}

	syncErrors := make([]error, 0, len(result.Errors))
	for _, loadErr := range result.Errors {
		syncErrors = append(syncErrors, fmt.Errorf("load movie %d: %s", loadErr.ID, loadErr.Message))
	}
	for _, movie := range result.Movies {
		if err := s.storage.SaveMovie(ctx, movie); err != nil {
			syncErrors = append(syncErrors, err)
		}
	}

	return errors.Join(syncErrors...)
}

func findMissingIDs(lastID int64, existingIDs []int64) []int64 {
	existing := make(map[int64]struct{}, len(existingIDs))
	for _, id := range existingIDs {
		existing[id] = struct{}{}
	}

	missing := make([]int64, 0)
	for id := int64(1); ; id++ {
		if _, ok := existing[id]; !ok {
			missing = append(missing, id)
		}
		if id == lastID {
			break
		}
	}
	return missing
}

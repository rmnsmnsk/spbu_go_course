package usecase

import (
	"context"
	"slices"
	"testing"

	"movie-service/internal/domain"
)

type fakeMovieStand struct {
	last      domain.Movie
	movies    map[int64]domain.Movie
	requested [][]int64
}

func (s *fakeMovieStand) GetLastMovie(context.Context) (domain.Movie, error) {
	return s.last, nil
}

func (s *fakeMovieStand) GetMovies(_ context.Context, ids []int64) (domain.LoadResult, error) {
	s.requested = append(s.requested, slices.Clone(ids))
	result := domain.LoadResult{Movies: make([]domain.Movie, 0, len(ids))}
	for _, id := range ids {
		result.Movies = append(result.Movies, s.movies[id])
	}
	return result, nil
}

type fakeMovieStorage struct {
	movies map[int64]domain.Movie
}

func (s *fakeMovieStorage) GetMovieIDs(context.Context, int64) ([]int64, error) {
	ids := make([]int64, 0, len(s.movies))
	for id := range s.movies {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids, nil
}

func (s *fakeMovieStorage) SaveMovie(_ context.Context, movie domain.Movie) error {
	s.movies[movie.ID] = movie
	return nil
}

func TestFindMissingIDs(t *testing.T) {
	tests := []struct {
		name     string
		lastID   int64
		existing []int64
		want     []int64
	}{
		{name: "empty database", lastID: 3, want: []int64{1, 2, 3}},
		{name: "gap in database", lastID: 4, existing: []int64{1, 3, 4}, want: []int64{2}},
		{name: "database is synchronized", lastID: 3, existing: []int64{1, 2, 3}, want: []int64{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findMissingIDs(tt.lastID, tt.existing)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("findMissingIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSynchronizerLoadsOnlyMissingMovies(t *testing.T) {
	stand := &fakeMovieStand{
		last: domain.Movie{ID: 4},
		movies: map[int64]domain.Movie{
			2: {ID: 2, Title: "two"},
			4: {ID: 4, Title: "four"},
		},
	}
	storage := &fakeMovieStorage{movies: map[int64]domain.Movie{
		1: {ID: 1, Title: "one"},
		3: {ID: 3, Title: "three"},
	}}
	synchronizer := NewSynchronizer(stand, storage)

	if err := synchronizer.Sync(context.Background()); err != nil {
		t.Fatalf("first Sync() error = %v", err)
	}
	if len(stand.requested) != 1 || !slices.Equal(stand.requested[0], []int64{2, 4}) {
		t.Fatalf("requested IDs = %v, want [[2 4]]", stand.requested)
	}
	if len(storage.movies) != 4 {
		t.Fatalf("stored movies = %d, want 4", len(storage.movies))
	}

	if err := synchronizer.Sync(context.Background()); err != nil {
		t.Fatalf("second Sync() error = %v", err)
	}
	if len(stand.requested) != 1 {
		t.Fatalf("second Sync() downloaded movies again: %v", stand.requested)
	}
}

package usecase

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"movie-service/internal/domain"
)

type MovieCatalog interface {
	GetAllMovies(ctx context.Context) ([]domain.Movie, error)
}

type MovieIndex interface {
	Replace(ctx context.Context, entries map[string][]int64) error
}

type Indexer struct {
	catalog MovieCatalog
	index   MovieIndex
}

func NewIndexer(catalog MovieCatalog, index MovieIndex) *Indexer {
	return &Indexer{catalog: catalog, index: index}
}

func (i *Indexer) Rebuild(ctx context.Context) error {
	movies, err := i.catalog.GetAllMovies(ctx)
	if err != nil {
		return fmt.Errorf("get movies for index: %w", err)
	}

	entries := make(map[string][]int64)
	for _, movie := range movies {
		movieWords := make(map[string]struct{})
		for _, word := range splitWords(movie.Title + " " + movie.Director) {
			movieWords[word] = struct{}{}
		}
		for word := range movieWords {
			entries[word] = append(entries[word], movie.ID)
		}
	}

	if err := i.index.Replace(ctx, entries); err != nil {
		return fmt.Errorf("replace movie index: %w", err)
	}
	return nil
}

func splitWords(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

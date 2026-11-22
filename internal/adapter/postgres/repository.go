package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"movie-service/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveMovie(ctx context.Context, movie domain.Movie) error {
	const query = `
		INSERT INTO movies (id, title, year, director)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			year = EXCLUDED.year,
			director = EXCLUDED.director`

	if _, err := r.db.ExecContext(ctx, query, movie.ID, movie.Title, movie.Year, movie.Director); err != nil {
		return fmt.Errorf("save movie %d: %w", movie.ID, err)
	}
	return nil
}

func (r *Repository) GetMovie(ctx context.Context, id int64) (domain.Movie, error) {
	const query = `
		SELECT id, title, year, director
		FROM movies
		WHERE id = $1`

	var movie domain.Movie
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.Title,
		&movie.Year,
		&movie.Director,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Movie{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Movie{}, fmt.Errorf("get movie %d: %w", id, err)
	}

	return movie, nil
}

func (r *Repository) GetAllMovies(ctx context.Context) ([]domain.Movie, error) {
	const query = `
		SELECT id, title, year, director
		FROM movies
		ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all movies: %w", err)
	}
	defer rows.Close()

	movies := make([]domain.Movie, 0)
	for rows.Next() {
		var movie domain.Movie
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.Year, &movie.Director); err != nil {
			return nil, fmt.Errorf("scan movie: %w", err)
		}
		movies = append(movies, movie)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read movies: %w", err)
	}
	return movies, nil
}

func (r *Repository) GetMoviesByIDs(ctx context.Context, ids []int64) ([]domain.Movie, error) {
	if len(ids) == 0 {
		return []domain.Movie{}, nil
	}

	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids))
	for position, id := range ids {
		placeholders = append(placeholders, fmt.Sprintf("$%d", position+1))
		args = append(args, id)
	}
	query := `
		SELECT id, title, year, director
		FROM movies
		WHERE id IN (` + strings.Join(placeholders, ", ") + `)
		ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get movies by IDs: %w", err)
	}
	defer rows.Close()

	movies := make([]domain.Movie, 0, len(ids))
	for rows.Next() {
		var movie domain.Movie
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.Year, &movie.Director); err != nil {
			return nil, fmt.Errorf("scan movie: %w", err)
		}
		movies = append(movies, movie)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read movies: %w", err)
	}
	return movies, nil
}

func (r *Repository) GetMoviesRange(ctx context.Context, from, to int64) (domain.LoadResult, error) {
	if from <= 0 || to <= 0 || from > to {
		return domain.LoadResult{}, domain.ErrInvalidRange
	}

	const query = `
		SELECT id, title, year, director
		FROM movies
		WHERE id BETWEEN $1 AND $2
		ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query, from, to)
	if err != nil {
		return domain.LoadResult{}, fmt.Errorf("get movies from %d to %d: %w", from, to, err)
	}
	defer rows.Close()

	result := domain.LoadResult{
		Movies: make([]domain.Movie, 0),
		Errors: make([]domain.LoadError, 0),
	}
	for rows.Next() {
		var movie domain.Movie
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.Year, &movie.Director); err != nil {
			return domain.LoadResult{}, fmt.Errorf("scan movie: %w", err)
		}
		result.Movies = append(result.Movies, movie)
	}
	if err := rows.Err(); err != nil {
		return domain.LoadResult{}, fmt.Errorf("read movies: %w", err)
	}

	movieIndex := 0
	for id := from; ; id++ {
		if movieIndex < len(result.Movies) && result.Movies[movieIndex].ID == id {
			movieIndex++
		} else {
			result.Errors = append(result.Errors, domain.LoadError{
				ID:      id,
				Message: domain.ErrNotFound.Error(),
			})
		}
		if id == to {
			break
		}
	}

	return result, nil
}

func (r *Repository) GetLastMovie(ctx context.Context) (domain.Movie, error) {
	const query = `
		SELECT id, title, year, director
		FROM movies
		ORDER BY id DESC
		LIMIT 1`

	var movie domain.Movie
	err := r.db.QueryRowContext(ctx, query).Scan(
		&movie.ID,
		&movie.Title,
		&movie.Year,
		&movie.Director,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Movie{}, domain.ErrNoMovies
	}
	if err != nil {
		return domain.Movie{}, fmt.Errorf("get last movie: %w", err)
	}

	return movie, nil
}

func (r *Repository) GetMovieIDs(ctx context.Context, to int64) ([]int64, error) {
	if to <= 0 {
		return nil, domain.ErrInvalidRange
	}

	const query = `
		SELECT id
		FROM movies
		WHERE id BETWEEN 1 AND $1
		ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query, to)
	if err != nil {
		return nil, fmt.Errorf("get movie IDs through %d: %w", to, err)
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan movie ID: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read movie IDs: %w", err)
	}

	return ids, nil
}

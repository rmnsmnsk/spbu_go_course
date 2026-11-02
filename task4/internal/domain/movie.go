package domain

import "errors"

var (
	ErrNotFound        = errors.New("movie not found")
	ErrInvalidRange    = errors.New("invalid movie range")
	ErrNoMovies        = errors.New("no movies found")
	ErrUpstreamTimeout = errors.New("movie stand request timed out")
)

type Movie struct {
	ID       int64
	Title    string
	Year     int32
	Director string
}

type LoadError struct {
	ID      int64
	Message string
}

type LoadResult struct {
	Movies []Movie
	Errors []LoadError
}

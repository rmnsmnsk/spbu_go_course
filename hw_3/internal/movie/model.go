package movie

import "errors"

var (
	ErrNotFound        = errors.New("movie not found")
	ErrInvalidRange    = errors.New("invalid movie range")
	ErrNoMovies        = errors.New("no movies found")
	ErrUpstreamTimeout = errors.New("movie stand request timed out")
)

type Movie struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Year     int32  `json:"year"`
	Director string `json:"director"`
}

type LoadError struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

type LoadResult struct {
	Movies []Movie     `json:"movies"`
	Errors []LoadError `json:"errors"`
}

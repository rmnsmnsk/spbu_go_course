package grpcapi

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	moviev1 "movie-service/gen/movie/v1"
	"movie-service/internal/movie"
)

type MovieService interface {
	LoadMovies(context.Context, int64, int64) (movie.LoadResult, error)
	GetLastMovie(context.Context) (movie.Movie, error)
}

type Server struct {
	moviev1.UnimplementedMovieServiceServer
	service MovieService
}

func New(service MovieService) *Server {
	return &Server{service: service}
}

func (s *Server) LoadMovies(ctx context.Context, req *moviev1.LoadMoviesRequest) (*moviev1.LoadMoviesResponse, error) {
	result, err := s.service.LoadMovies(ctx, req.GetFrom(), req.GetTo())
	if err != nil {
		return nil, grpcError(err)
	}

	response := &moviev1.LoadMoviesResponse{
		Movies: make([]*moviev1.Movie, 0, len(result.Movies)),
		Errors: make([]*moviev1.MovieError, 0, len(result.Errors)),
	}
	for _, loaded := range result.Movies {
		response.Movies = append(response.Movies, toProtoMovie(loaded))
	}
	for _, loadErr := range result.Errors {
		response.Errors = append(response.Errors, &moviev1.MovieError{Id: loadErr.ID, Message: loadErr.Message})
	}
	return response, nil
}

func (s *Server) GetLastMovie(ctx context.Context, _ *moviev1.GetLastMovieRequest) (*moviev1.Movie, error) {
	result, err := s.service.GetLastMovie(ctx)
	if err != nil {
		return nil, grpcError(err)
	}
	return toProtoMovie(result), nil
}

func toProtoMovie(value movie.Movie) *moviev1.Movie {
	return &moviev1.Movie{
		Id:       value.ID,
		Title:    value.Title,
		Year:     value.Year,
		Director: value.Director,
	}
}

func grpcError(err error) error {
	switch {
	case errors.Is(err, movie.ErrInvalidRange):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, movie.ErrNoMovies), errors.Is(err, movie.ErrNotFound):
		return status.Error(codes.NotFound, "movie not found")
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, movie.ErrUpstreamTimeout):
		return status.Error(codes.DeadlineExceeded, "movie stand request timed out")
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	default:
		return status.Error(codes.Unavailable, err.Error())
	}
}

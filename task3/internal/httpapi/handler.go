package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"movie-service/api"
	"movie-service/internal/movie"
)

type MovieService interface {
	LoadMovies(context.Context, int64, int64) (movie.LoadResult, error)
	GetLastMovie(context.Context) (movie.Movie, error)
}

type Handler struct {
	service MovieService
}

func New(service MovieService) http.Handler {
	h := &Handler{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /load-movies", h.loadMovies)
	mux.HandleFunc("GET /last-movie", h.lastMovie)
	mux.HandleFunc("GET /swagger/openapi.yaml", serveOpenAPI)
	mux.Handle("/swagger/", httpSwagger.Handler(httpSwagger.URL("/swagger/openapi.yaml")))
	return mux
}

func (h *Handler) loadMovies(w http.ResponseWriter, r *http.Request) {
	from, err := requiredPositiveInt64(r, "from")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	to, err := requiredPositiveInt64(r, "to")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.LoadMovies(r.Context(), from, to)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	if len(result.Errors) == 0 {
		writeJSON(w, http.StatusOK, result.Movies)
		return
	}
	writeJSON(w, http.StatusMultiStatus, result)
}

func (h *Handler) lastMovie(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.GetLastMovie(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func requiredPositiveInt64(r *http.Request, name string) (int64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return 0, errors.New("query parameter " + name + " is required")
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, errors.New("query parameter " + name + " must be a positive integer")
	}
	return value, nil
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, movie.ErrInvalidRange):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, movie.ErrNoMovies), errors.Is(err, movie.ErrNotFound):
		writeError(w, http.StatusNotFound, "movie not found")
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, movie.ErrUpstreamTimeout):
		writeError(w, http.StatusGatewayTimeout, "movie stand request timed out")
	case errors.Is(err, context.Canceled):
		writeError(w, http.StatusRequestTimeout, "request canceled")
	default:
		writeError(w, http.StatusBadGateway, err.Error())
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func serveOpenAPI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	_, _ = w.Write(api.OpenAPISpec)
}

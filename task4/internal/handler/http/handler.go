package httphandler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"movie-service/api"
	"movie-service/internal/domain"
	"movie-service/internal/usecase"
)

type Handler struct {
	service usecase.MovieUseCase
}

type movieResponse struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Year     int32  `json:"year"`
	Director string `json:"director"`
}

type loadErrorResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

type loadMoviesResponse struct {
	Movies []movieResponse     `json:"movies"`
	Errors []loadErrorResponse `json:"errors"`
}

func New(service usecase.MovieUseCase) http.Handler {
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

	response := toLoadMoviesResponse(result)
	if len(response.Errors) == 0 {
		writeJSON(w, http.StatusOK, response.Movies)
		return
	}
	writeJSON(w, http.StatusMultiStatus, response)
}

func (h *Handler) lastMovie(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.GetLastMovie(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toMovieResponse(result))
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

func toLoadMoviesResponse(result domain.LoadResult) loadMoviesResponse {
	response := loadMoviesResponse{
		Movies: make([]movieResponse, 0, len(result.Movies)),
		Errors: make([]loadErrorResponse, 0, len(result.Errors)),
	}
	for _, loaded := range result.Movies {
		response.Movies = append(response.Movies, toMovieResponse(loaded))
	}
	for _, loadErr := range result.Errors {
		response.Errors = append(response.Errors, loadErrorResponse{ID: loadErr.ID, Message: loadErr.Message})
	}
	return response
}

func toMovieResponse(value domain.Movie) movieResponse {
	return movieResponse{
		ID:       value.ID,
		Title:    value.Title,
		Year:     value.Year,
		Director: value.Director,
	}
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidRange):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrNoMovies), errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "movie not found")
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, domain.ErrUpstreamTimeout):
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

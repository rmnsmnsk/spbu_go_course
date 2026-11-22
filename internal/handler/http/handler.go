package httphandler

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"movie-service/internal/usecase"
)

type Handler struct {
	service usecase.MovieUseCase
}

func New(service usecase.MovieUseCase) http.Handler {
	h := &Handler{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /load-movies", h.loadMovies)
	mux.HandleFunc("GET /last-movie", h.lastMovie)
	mux.HandleFunc("GET /search", h.searchMovies)
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

func (h *Handler) searchMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := h.service.SearchMovies(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toMoviesResponse(movies))
}

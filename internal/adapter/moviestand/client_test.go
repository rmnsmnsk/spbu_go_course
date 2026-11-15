package moviestand

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func newTestStand(t *testing.T, lastID int64) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 2 || parts[1] != "info.0.json" {
			http.NotFound(w, r)
			return
		}
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil || id < 1 || id > lastID {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(movieResponse{
			ID:       id,
			Title:    "movie " + parts[0],
			Year:     2000 + int32(id),
			Director: "director",
		})
	}))
}

func TestGetLastMovie(t *testing.T) {
	server := newTestStand(t, 3)
	defer server.Close()
	client := New(server.URL, time.Second, 2)

	movie, err := client.GetLastMovie(context.Background())
	if err != nil {
		t.Fatalf("GetLastMovie() error = %v", err)
	}
	if movie.ID != 3 {
		t.Fatalf("GetLastMovie().ID = %d, want 3", movie.ID)
	}
}

func TestGetMoviesReturnsIndividualErrors(t *testing.T) {
	server := newTestStand(t, 3)
	defer server.Close()
	client := New(server.URL, time.Second, 2)

	result, err := client.GetMovies(context.Background(), []int64{1, 4})
	if err != nil {
		t.Fatalf("GetMovies() error = %v", err)
	}
	if len(result.Movies) != 1 || result.Movies[0].ID != 1 {
		t.Fatalf("GetMovies().Movies = %v, want movie 1", result.Movies)
	}
	if len(result.Errors) != 1 || result.Errors[0].ID != 4 {
		t.Fatalf("GetMovies().Errors = %v, want error for movie 4", result.Errors)
	}
}

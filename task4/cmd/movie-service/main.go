package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"

	moviev1 "movie-service/gen/movie/v1"
	"movie-service/internal/adapter/moviestand"
	grpchandler "movie-service/internal/handler/grpc"
	httphandler "movie-service/internal/handler/http"
	"movie-service/internal/usecase"
)

const (
	grpcAddr        = ":9090"
	defaultBaseURL  = "https://homeworksite.site"
	defaultHTTPAddr = ":8080"
	requestTimeout  = 5 * time.Second
	workerCount     = 10
)

func main() {
	baseURL := envOrDefault("MOVIE_BASE_URL", defaultBaseURL)
	httpAddr := envOrDefault("HTTP_ADDR", defaultHTTPAddr)
	stand := moviestand.New(baseURL, requestTimeout)
	movieService := usecase.New(stand, workerCount)

	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	moviev1.RegisterMovieServiceServer(grpcServer, grpchandler.New(movieService))
	go func() {
		log.Printf("gRPC server is listening on %s", grpcAddr)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("HTTP server is listening on %s", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, httphandler.New(movieService)))
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

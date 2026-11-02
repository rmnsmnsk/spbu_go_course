package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"

	moviev1 "movie-service/gen/movie/v1"
	"movie-service/internal/grpcapi"
	"movie-service/internal/httpapi"
	"movie-service/internal/service"
	"movie-service/internal/standclient"
)

const grpcAddr = ":9090"

func main() {
	baseURL := envOrDefault("MOVIE_BASE_URL", "https://homeworksite.site")
	httpAddr := envOrDefault("HTTP_ADDR", ":8080")
	stand := standclient.New(baseURL, 5*time.Second)
	movieService := service.New(stand, 10)

	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	moviev1.RegisterMovieServiceServer(grpcServer, grpcapi.New(movieService))
	go func() {
		log.Printf("gRPC server is listening on %s", grpcAddr)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("HTTP server is listening on %s", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, httpapi.New(movieService)))
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

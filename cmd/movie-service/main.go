package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	moviev1 "movie-service/gen/movie/v1"
	"movie-service/internal/adapter/moviestand"
	"movie-service/internal/adapter/postgres"
	redisadapter "movie-service/internal/adapter/redis"
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
	shutdownTimeout = 5 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	baseURL := envOrDefault("MOVIE_BASE_URL", defaultBaseURL)
	httpAddr := envOrDefault("HTTP_ADDR", defaultHTTPAddr)
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	redisAddress := os.Getenv("REDIS_ADDR")
	if redisAddress == "" {
		return errors.New("REDIS_ADDR is required")
	}

	db, err := postgres.Open(databaseURL)
	if err != nil {
		return fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	defer db.Close()
	if err := postgres.Migrate(db); err != nil {
		return fmt.Errorf("migrate PostgreSQL: %w", err)
	}
	redisClient, err := redisadapter.Open(ctx, redisAddress)
	if err != nil {
		return fmt.Errorf("connect to Redis: %w", err)
	}
	defer redisClient.Close()

	repository := postgres.NewRepository(db)
	stand := moviestand.New(baseURL, requestTimeout, workerCount)
	synchronizer := usecase.NewSynchronizer(stand, repository)
	log.Printf("synchronizing movies with %s", baseURL)
	if err := synchronizer.Sync(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return fmt.Errorf("synchronize movies: %w", err)
	}
	log.Print("movie synchronization completed")
	index := redisadapter.NewIndex(redisClient)
	indexer := usecase.NewIndexer(repository, index)
	log.Print("building movie search index")
	if err := indexer.Rebuild(ctx); err != nil {
		return fmt.Errorf("build movie search index: %w", err)
	}
	log.Print("movie search index completed")

	movieService := usecase.New(repository, index)

	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("listen for gRPC: %w", err)
	}
	defer grpcListener.Close()

	grpcServer := grpc.NewServer()
	moviev1.RegisterMovieServiceServer(grpcServer, grpchandler.New(movieService))
	serverErrors := make(chan error, 2)
	go func() {
		log.Printf("gRPC server is listening on %s", grpcAddr)
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			serverErrors <- fmt.Errorf("serve gRPC: %w", err)
		}
	}()

	httpServer := &http.Server{
		Addr:              httpAddr,
		Handler:           httphandler.New(movieService),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Printf("HTTP server is listening on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("serve HTTP: %w", err)
		}
	}()

	select {
	case err := <-serverErrors:
		grpcServer.Stop()
		_ = httpServer.Close()
		return err
	case <-ctx.Done():
	}

	log.Print("shutting down servers")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	grpcServer.GracefulStop()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}
	return nil
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

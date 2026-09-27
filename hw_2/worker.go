package main

import (
	"context"
	"sync"
	"time"
)

func worker(
	ctx context.Context,
	jobs <-chan int,
	results chan<- Result,
	timeout time.Duration,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for i := range jobs {
		if ctx.Err() != nil {
			return
		}

		body, err := fetchFilm(ctx, i, timeout)

		if err != nil {
			if ctx.Err() != nil {
				return
			}
			results <- Result{
				ID:  i,
				Err: err,
			}
			continue
		}

		movie, err := parseJson(body)

		if err != nil {
			results <- Result{
				ID:  i,
				Err: err,
			}
			continue
		}

		results <- Result{
			ID:    i,
			Movie: movie,
		}

	}
}

func sendJobs(ctx context.Context, jobs chan<- int, from int, to int) {
	defer close(jobs)
	for i := from; i <= to; i++ {
		select {
		case jobs <- i:
		case <-ctx.Done():
			return
		}

	}
}

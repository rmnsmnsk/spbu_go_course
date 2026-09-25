package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Result struct {
	ID    int
	Movie *Movie
	Err   error
}

func main() {

	from := flag.Int("from", 0, "id первого фильма")
	to := flag.Int("to", 0, "id последнего фильма")
	workers := flag.Int("workers", 10, "сколько воркеров")
	timeout := flag.Duration("timeout", 5*time.Second, "сколько таймаут HTTP запроса")

	flag.Parse()

	if *from == 0 {
		fmt.Println("flag --from is required")
		os.Exit(1)
	}

	if *to == 0 {
		fmt.Println("flag --to is required")
		os.Exit(1)
	}

	if *from > *to {
		fmt.Println("flag --from can't be bigger than flag --to")
		os.Exit(1)
	}

	if *workers <= 0 {
		fmt.Println("flag --workers can't be <= 0")
		os.Exit(1)
	}

	if *timeout <= 0 {
		fmt.Println("flag --timeout can't be < 0")
		os.Exit(1)
	}

	/*fmt.Println("from:", *from)
	fmt.Println("to:", *to)
	fmt.Println("workers:", *workers)
	fmt.Println("timeout:", *timeout)*/

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
	)
	defer stop()

	jobs := make(chan int)
	results := make(chan Result)

	var wg sync.WaitGroup

	for w := 0; w < *workers; w++ {
		wg.Add(1)
		go worker(ctx, jobs, results, *timeout, &wg)
	}

	go sendJobs(ctx, jobs, *from, *to)

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {

		if result.Err != nil {
			fmt.Printf("movie %d error: %v\n", result.ID, result.Err)
			continue
		}

		movie := result.Movie

		fmt.Printf("%d — %s — %d — %s\n",
			movie.ID,
			movie.Title,
			movie.Year,
			movie.Director,
		)

	}

	if ctx.Err() != nil {
		fmt.Println("operation interrupted")
	}

}

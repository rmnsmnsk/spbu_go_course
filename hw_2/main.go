package main

import (
	"fmt"
	//"io"
	"flag"
	"time"
	"os"
)

func main(){

	from := flag.Int("from", 0, "id первого фильма")
	to := flag.Int("to", 0, "id последнего фильма")
	workers := flag.Int("workers", 10, "сколько воркеров")
	timeout := flag.Duration("timeout", 5 * time.Second, "сколько таймаут HTTP запроса")

	flag.Parse()

	if (*from == 0){
		fmt.Println("flag --from is required")
		os.Exit(1)
	}

	if (*to == 0){
		fmt.Println("flag --to is requested")
		os.Exit(1)
	}

	if (*from > *to){
		fmt.Println("flag --from can't be bigger than flag --to")
		os.Exit(1)
	}

	if (*workers <= 0){
		fmt.Println("flag --workers can't be <= 0")
		os.Exit(1)
	}

	if (*timeout < 0){
		fmt.Println("flag --timeout can't be < 0")
		os.Exit(1)
	}

	/*fmt.Println("from:", *from)
	fmt.Println("to:", *to)
	fmt.Println("workers:", *workers)
	fmt.Println("timeout:", *timeout)*/

	for i := *from; i <= *to; i++{

		body, err := fetchFilm(i)

		if err != nil{
			fmt.Println("movie error", err)
			continue
		}

		movie, err := parseJson(body)

		if err != nil{
			fmt.Println("parsing error", err)
			continue
		}

		fmt.Printf("%d — %s — %d — %s\n",
			movie.ID,
			movie.Title,
			movie.Year,
			movie.Director,
		)

		fmt.Print(string(body))

	}

}
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
		return
		os.Exit(1)
	}

	if (*from > *to){
		fmt.Println("flag --from can't be bigger than flag --to")
		return
		os.Exit(1)
	}

	if (*workers <= 0){
		fmt.Println("flag --workers can't be <= 0")
		return
		os.Exit(1)
	}

	if (*timeout < 0){
		fmt.Println("flag --timeout can't be < 0")
		return
		os.Exit(1)
	}

	/*fmt.Println("from:", *from)
	fmt.Println("to:", *to)
	fmt.Println("workers:", *workers)
	fmt.Println("timeout:", *timeout)*/

	for i := *from; i <= *to; i++{

		body, err := fetchFilm(i)

		if err != nil{
			return
		}

		movie, err := parseJson(body)

		fmt.Println(movie)

		fmt.Print(string(body))

	}

}
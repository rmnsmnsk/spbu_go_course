package main

import (
	"fmt"
	"io"
	"net/http"
	"flag"
	"time"
)

func main(){

	from := flag.Int("from", 0, "id первого фильма")
	to := flag.Int("to", 0, "id последнего фильма")
	workers := flag.Int("workers", 0, "сколько воркеров")
	timeout := flag.Duration("timeout", 5 * time.Second, "сколько таймаут HTTP запроса")

	flag.Parse()

	fmt.Println("from:", *from)
	fmt.Println("to:", *to)
	fmt.Println("workers:", *workers)
	fmt.Println("timeout:", *timeout)

	resp, err := http.Get("https://homeworksite.site/1/info.0.json");

	if err != nil{
		fmt.Println("request error:", err)
		return
	}

	body, err1 := io.ReadAll(resp.Body);

	if err1 != nil{
		fmt.Println("reading error:", err1)
		return
	}

	movie, err := parseJson(body)
	fmt.Println(movie)

	defer resp.Body.Close()

	fmt.Print(string(body))

	fmt.Println("status:", resp.Status)

}
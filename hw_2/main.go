package main

import (
	"fmt"
	"io"
	"net/http"
)

func main(){

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
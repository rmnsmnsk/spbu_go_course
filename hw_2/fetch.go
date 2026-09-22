package main

import(
	"net/http"
	"io"
	"fmt"

)



func fetchFilm(i int)([]byte, error){

	url := fmt.Sprintf("https://homeworksite.site/%d/info.0.json", i)

	resp, err := http.Get(url)

	if err != nil{
		fmt.Println("request error:", err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil{
		fmt.Println("reading error", err)
		return nil, err
	}

	return body, nil

}
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

	if (resp.StatusCode < 200 || resp.StatusCode > 299){
		return nil, fmt.Errorf("movie %d: HTTP status %d", i, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil{
		return nil, err
	}

	return body, nil

}
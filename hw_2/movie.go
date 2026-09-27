package main

import "encoding/json"

type Movie struct {
	ID       int
	Title    string
	Year     int
	Director string
}

func parseJson(data []byte) (*Movie, error) {

	var movie Movie
	err := json.Unmarshal(data, &movie)

	if err != nil {
		return nil, err
	}

	return &movie, nil
}

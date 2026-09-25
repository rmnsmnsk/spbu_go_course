package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

func fetchFilm(ctx context.Context, i int, timeout time.Duration) ([]byte, error) {

	requestCtx, cancel := context.WithTimeout(ctx, timeout)

	defer cancel()

	url := fmt.Sprintf("https://homeworksite.site/%d/info.0.json", i)

	req, err := http.NewRequestWithContext(
		requestCtx,
		http.MethodGet,
		url,
		nil,
	)

	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("movie %d: HTTP status %d", i, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil

}

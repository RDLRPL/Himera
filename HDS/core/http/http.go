package http

import (
	"io"
	"log"
	"net/http"
)

type Response struct {
	UserAgent string
	Page      string
	Done      bool
}

func GETRequest(adress string, Ua string) (*Response, error) {

	req, err := http.NewRequest("GET", adress, nil)
	if err != nil {
		log.Printf("HTTP request creation error: %v", err)
		return nil, err
	}

	req.Header.Set("User-Agent", Ua)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("HTTP request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Read error: %v", err)
		return nil, err
	}
	htmlString := string(bodyBytes)

	return &Response{
		UserAgent: Ua,
		Page:      htmlString,
		Done:      true,
	}, nil
}

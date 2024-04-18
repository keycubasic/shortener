// Package handler show examples
package handler

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/keycubasic/shortener/internal/dto"
)

// ExampleURLHandler_pingDB demonstrates calling the pingDB handler endpoint.
func ExampleURLHandler_pingDB() {
	res, err := http.Get("http://localhost:8080/ping")
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d", res.StatusCode)
	}

	log.Print("Ping success")
}

// ExampleURLHandler_getOriginalURL demonstrates calling the getOriginalURL handler endpoint.
func ExampleURLHandler_getOriginalURL() {
	res, err := http.Get("http://localhost:8080/TgIlZU3m")
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("getOriginalURL body: %s", body)
}

// ExampleURLHandler_shortenURL demonstrates calling the shortenURL handler endpoint.
func ExampleURLHandler_shortenURL() {
	res, err := http.Post("http://localhost:8080", "text/plain; charset=utf-8", strings.NewReader("https://ya.ru"))
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("shortenURL body: %s", body)
}

// ExampleURLHandler_shortenURLJSON demonstrates calling the shortenURLJSON handler endpoint.
func ExampleURLHandler_shortenURLJSON() {
	res, err := http.Post("http://localhost:8080/api/shorten", "application/json",
		strings.NewReader("{\"url\":\"http://yahoo.com\"}"))
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}

	response := new(dto.Response)
	if err := json.Unmarshal(body, &response); err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("shortenURLJSON response: %s", response)
}

// ExampleURLHandler_shortenURLJSONBatch demonstrates calling the shortenURLJSONBatch handler endpoint.
func ExampleURLHandler_shortenURLJSONBatch() {
	res, err := http.Post("http://localhost:8080/api/shorten/batch", "application/json", strings.NewReader(
		`[
			{
				"correlation_id": "3",
				"original_url": "https://yandex.ru"
			},
			{
				"correlation_id": "3",
				"original_url": "https://google.com"
			}
		] `))
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}

	resLinksDto := make(dto.LinksDto, 0)
	if err := json.Unmarshal(body, &resLinksDto); err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("shortenURLJSONBatch response: %s", resLinksDto)
}

// ExampleURLHandler_getUserURLs demonstrates calling the getUserURLs handler endpoint.
func ExampleURLHandler_getUserURLs() {
	res, err := http.Get("http://localhost:8080/api/user/urls")
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}

	resLinksDto := make(dto.LinksDto, 0)
	if err := json.Unmarshal(body, &resLinksDto); err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("getUserURLs result: %s", resLinksDto)
}

// ExampleURLHandler_deleteUserURLs demonstrates calling the deleteUserURLs handler endpoint.
func ExampleURLHandler_deleteUserURLs() {
	r, err := http.NewRequest("DELETE", "localhost:8080/api/user/urls",
		strings.NewReader("[\"6YGS4ZUF\",\"rntA0dVu\",\"Yi66Tx-0\",\"s7L9TR66\"]"))
	if err != nil {
		log.Fatal(err)
	}

	r.RequestURI = ""
	r.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d", res.StatusCode)
	}

	log.Print("deleteUserURLs success")
}

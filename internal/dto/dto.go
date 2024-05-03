// Package dto contains data transfer objects.
package dto

// Request stores the request object.
// @Description URL shortening request in JSON format
type Request struct {
	URL string `json:"url"`
}

// Response stores the response object.
// @Description URL shortening response in JSON format
type Response struct {
	SortenURL string `json:"result"`
}

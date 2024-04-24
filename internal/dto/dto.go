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

// NewRequest creates a new Request object
func NewRequest() *Request {
	return new(Request)
}

// NewResponse creates a new Response object
func NewResponse(shorten string) *Response {
	return &Response{SortenURL: shorten}
}

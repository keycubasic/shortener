// Package dto contains data transfer objects.
package dto

// Request stores the request object.
type Request struct {
	URL string `json:"url"`
}

// Response stores the response object.
type Response struct {
	SortenURL string `json:"result"`
}

func NewRequest() *Request {
	return new(Request)
}

func NewResponse(shorten string) *Response {
	return &Response{SortenURL: shorten}
}

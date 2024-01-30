package dto

// go:generate easyjson -all dto.go
type Request struct {
	URL string `json:"url"`
}

type Response struct {
	SortenURL string `json:"result"`
}

func NewRequest() *Request {
	return new(Request)
}

func NewResponse(shorten string) *Response {
	return &Response{SortenURL: shorten}
}

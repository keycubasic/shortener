package dto

import (
	"encoding/json"
	"io"
	"net/url"

	"github.com/asaskevich/govalidator"
)

type Request struct {
	URL string `json:"url" valid:"url"`
}

type Response struct {
	SortenURL string `json:"result"`
}

func NewRequest() *Request {
	return new(Request)
}

func (req *Request) Unmarshal(r io.Reader) error {
	return json.NewDecoder(r).Decode(req)
}

func (req *Request) ValidateRequest() (bool, error) {

	if ok, err := govalidator.ValidateStruct(req); !ok {
		return ok, err
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		return false, err
	}

	return true, nil
}

func NewResponse(shorten string) *Response {
	return &Response{SortenURL: shorten}
}

func (res *Response) Marshal() ([]byte, error) {
	r, err := json.Marshal(res)
	return r, err
}

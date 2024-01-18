package dto

type Request struct {
	URL string `json:"url" valid:"url"`
}

type Response struct {
	SortenURL string `json:"result"`
}
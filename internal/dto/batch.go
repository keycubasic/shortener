package dto

//easyjson:json
type LinksDto []LinkDto

//easyjson:json
type LinkDto struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	OriginalURL   string `json:"original_url,omitempty"`
	ShortURL      string `json:"short_url,omitempty"`
}

func NewLinkDto() *LinkDto {
	return new(LinkDto)
}

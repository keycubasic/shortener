// Package dto contains data transfer objects for batch processing.
package dto

// Slice stores a slice of links for batch processing.
type Slice []string

// LinksDto stores a slice of data transfer objects.
type LinksDto []LinkDto

// LinkDto stores the data transfer object.
// @Description Data transfer object for batch mode
type LinkDto struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	OriginalURL   string `json:"original_url,omitempty"`
	ShortURL      string `json:"short_url,omitempty"`
}

// NewLinkDto creates a new LinkDto object
func NewLinkDto() *LinkDto {
	return new(LinkDto)
}

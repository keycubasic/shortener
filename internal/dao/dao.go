package dao

// go:generate easyjson -all dao.go
type LinksDao []*LinkDao
type LinkDao struct {
	UUID          int    `json:"uuid"`
	ShortURL      string `json:"short_url,omitempty"`
	OriginalURL   string `json:"original_url,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
}

func NewLinksDao() LinksDao {
	return []*LinkDao{}
}

func NewLinkDao() *LinkDao {
	return new(LinkDao)
}

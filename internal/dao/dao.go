package dao

// go:generate easyjson -all dao.go
type LinksDao []*LinkDao
type LinkDao struct {
	UUID          int    `json:"uuid"`
	ShortURL      string `json:"short_url"`
	OriginalURL   string `json:"original_url"`
	CorrelationID string
}

func NewLinksDao() LinksDao {
	return []*LinkDao{}
}

func NewLinkDao() *LinkDao {
	return new(LinkDao)
}

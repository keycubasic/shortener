package dao

// go:generate easyjson -all dao.go
type LinkDao struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewLinkDao() *LinkDao {
	return new(LinkDao)
}

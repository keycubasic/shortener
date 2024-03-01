package dao

// go:generate easyjson -all dao.go
type LinksDao []*LinkDao

type LinksDaoBatch []*LinkDaoBatch

type LinkDao struct {
	UUID          int    `json:"uuid"`
	ShortURL      string `json:"short_url" db:"short_url"`
	OriginalURL   string `json:"original_url" db:"original_url"`
	CorrelationID string `json:"-"`
	UserID        int    `json:"-" db:"user_id"`
	DeletedFlag   bool   `json:"-"`
}

type LinkDaoBatch struct {
	ShortURL    string
	OriginalURL string
	DeletedFlag bool
}

func NewLinkDao() *LinkDao {
	return new(LinkDao)
}

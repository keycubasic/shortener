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
	DeletedFlag   bool   `json:"-" db:"is_deleted"`
}

type LinkDaoBatch struct {
	ShortURL      string `db:"short_url"`
	OriginalURL   string `db:"original_url"`
	DeletedFlag   bool   `db:"is_deleted"`
}

func NewLinkDao() *LinkDao {
	return new(LinkDao)
}

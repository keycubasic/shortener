// Package dao contains data access objects.
package dao

// LinksDao stores a slice of data access objects.
type LinksDao []*LinkDao

// LinksDaoBatch stores a slice of batch processing data access objects.
type LinksDaoBatch []*LinkDaoBatch

// LinkDao stores a data access object.
type LinkDao struct {
	UUID          int    `json:"uuid"`
	ShortURL      string `json:"short_url" db:"short_url"`
	OriginalURL   string `json:"original_url" db:"original_url"`
	CorrelationID string `json:"-"`
	UserID        int    `json:"-" db:"user_id"`
	DeletedFlag   bool   `json:"-"`
}

// LinkDaoBatch stores a batch data access object.
type LinkDaoBatch struct {
	ShortURL    string
	OriginalURL string
	DeletedFlag bool
}

// Stats stores the statistics object.
type Stats struct {
	URLs  int
	Users int
}

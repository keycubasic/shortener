package dao

import (
	"encoding/json"
)

type LinkDao struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewLinkDao() *LinkDao {
	return &LinkDao{}
}

func (link *LinkDao) Marshal() ([]byte, error) {
	return json.Marshal(link)
}

func (link *LinkDao) Unmarshal(data []byte) error {
	return json.Unmarshal(data, link)
}

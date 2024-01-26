package app

import (
	"crypto/md5"
	"encoding/base64"
	"regexp"

	"github.com/keycubasic/shortener/internal/storage"
)

type configuration interface {
	GetFilePath() string
}

type App struct {
	db *storage.Links
}

func NewApp(cfg configuration) *App {
	db := storage.NewLinks()
	db.Load(cfg)
	return &App{db: db}
}

func (a *App) GetLink(id string) (string, bool) {
	return a.db.GetLink(id)
}

func (a *App) Count() int {
	return a.db.Count()
}

func (a *App) CreateShortLink(link string) (string, error) {
	key := genKey(link)
	a.db.SetLink(key, link)
	return key, nil
}

func genKey(link string) string {
	md5checksum := md5.Sum([]byte(link))
	base64encoded := base64.URLEncoding.EncodeToString([]byte(md5checksum[:]))
	re := regexp.MustCompile("[+/=_]")
	base64encoded = re.ReplaceAllString(base64encoded, "")
	key := base64encoded[:8]
	return key
}

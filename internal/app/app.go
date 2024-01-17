package app

import (
	"crypto/md5"
	"encoding/base64"
	"regexp"
)

type database interface {
	GetLink(key string) (string, bool)
	SetLink(key string, link string)
}

type App struct {
	db database
}

func NewApp(db database) *App {
	return &App{db: db}
}

func (a *App) GetLink(id string) (string, bool) {
	return a.db.GetLink(id)
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

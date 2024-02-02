package app

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"regexp"
	"time"
)

//go:generate mockgen -source app.go -destination mock_storage.go -package mock database
type DB interface {
	GetLink(key string) (string, error)
	SetLink(key string, link string) error
	Close() error
	PingContext(ctx context.Context) error
}

type App struct {
	db DB
}

func NewApp(db DB) *App {
	return &App{
		db: db,
	}
}

func (a *App) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.db.PingContext(ctx)
}

func (a *App) GetLink(id string) (string, error) {
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

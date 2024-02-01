package app

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"regexp"
	"time"
)

type databaseLegacy interface {
	GetLink(key string) (string, bool)
	SetLink(key string, link string) error
}

type App struct {
	db       *sql.DB
	dbLegacy databaseLegacy
}

func NewApp(dbLegacy databaseLegacy, db *sql.DB) *App {
	return &App{
		db:       db,
		dbLegacy: dbLegacy,
	}
}

func (a *App) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return a.db.PingContext(ctx)
}

func (a *App) GetLink(id string) (string, bool) {
	return a.dbLegacy.GetLink(id)
}

func (a *App) CreateShortLink(link string) (string, error) {
	key := genKey(link)
	a.dbLegacy.SetLink(key, link)
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

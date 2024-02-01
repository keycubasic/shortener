package app

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"fmt"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

type configuration interface {
	GetDatabaseDSN() string
}

type databaseLegacy interface {
	GetLink(key string) (string, bool)
	SetLink(key string, link string) error
}

type App struct {
	cfg      configuration
	db       *sql.DB
	dbLegacy databaseLegacy
}

func NewApp(cfg configuration, dbLegacy databaseLegacy) *App {
	return &App{
		cfg:      cfg,
		dbLegacy: dbLegacy,
	}
}

func (a *App) OpenConnection() error {
	log.Infoln("Database DSN:", a.cfg.GetDatabaseDSN())
	db, err := sql.Open("pgx", a.cfg.GetDatabaseDSN())

	if err != nil {
		return fmt.Errorf("database open connection failed: %w", err)
	}

	log.Infoln("Database connected")
	a.db = db
	return nil
}

func (a *App) CloseConnection() error {
	return a.db.Close()
}

func (a *App) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Cfg interface {
	GetDatabaseDSN() string
}

type Links struct {
	sync.RWMutex
	db *sql.DB
}

func NewLinks(cfg Cfg) (*Links, error) {
	log.Infoln("Database DSN:", cfg.GetDatabaseDSN())
	db, err := sql.Open("pgx", cfg.GetDatabaseDSN())

	if err != nil {
		return nil, fmt.Errorf("database open connection failed: %w", err)
	}

	log.Infoln("Database connected...")

	_, err = db.ExecContext(context.Background(),
		"CREATE TABLE IF NOT EXISTS links (key TEXT primary key, value TEXT); GRANT ALL PRIVILEGES ON links TO PUBLIC;")

	if err != nil {
		return nil, fmt.Errorf("create table links failed: %w", err)
	}

	return &Links{db: db}, nil
}

func (ls *Links) GetLink(key string) (string, error) {
	ls.RLock()
	defer ls.RUnlock()

	row := ls.db.QueryRowContext(context.Background(),
		"SELECT value FROM links WHERE key = $1", key)
	var val string

	err := row.Scan(&val)

	if err != nil {
		return "", fmt.Errorf("select link from database failed: %w", err)
	}

	return val, nil
}

func (ls *Links) SetLink(key string, link string) error {
	if _, err := ls.GetLink(key); err != nil {
		ls.Lock()
		defer ls.Unlock()

		_, err := ls.db.ExecContext(context.Background(),
			"INSERT INTO links (key, value) VALUES ($1,$2)", key, link)

		if err != nil {
			return fmt.Errorf("insert link to database failed: %w", err)
		}
	}

	return nil
}

func (ls *Links) Count() (int, error) {
	//unused
	row := ls.db.QueryRowContext(context.Background(),
		"SELECT COUNT(1) FROM links")
	var cnt int

	err := row.Scan(&cnt)

	if err != nil {
		return 0, fmt.Errorf("select links count from database failed: %w", err)
	}

	return cnt, nil
}

func (ls *Links) Close() error {
	return ls.db.Close()
}

func (ls *Links) PingContext(ctx context.Context) error {
	return ls.db.PingContext(ctx)
}

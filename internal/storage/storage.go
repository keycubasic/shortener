package storage

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"sync"

	"github.com/keycubasic/shortener/internal/dao"
	"github.com/keycubasic/shortener/internal/gen"
	log "github.com/sirupsen/logrus"
)

type Cfg interface {
	GetDatabaseDSN() string
}

type Links struct {
	sync.RWMutex
	db *sql.DB
}

func NewLinks(ctx context.Context, cfg Cfg) (*Links, error) {
	log.Infoln("Database DSN:", cfg.GetDatabaseDSN())
	db, err := sql.Open("pgx", cfg.GetDatabaseDSN())

	if err != nil {
		return nil, fmt.Errorf("database open connection failed: %w", err)
	}

	log.Infoln("Database connected...")

	_, err = db.ExecContext(ctx,
		"CREATE TABLE IF NOT EXISTS links( "+
			"uuid int PRIMARY KEY GENERATED ALWAYS AS IDENTITY, "+
			"short_url VARCHAR(8), "+
			"original_url VARCHAR(100) "+
			");")

	if err != nil {
		return nil, fmt.Errorf("create table links failed: %w", err)
	}

	_, err = db.ExecContext(ctx,
		"CREATE INDEX IF NOT EXISTS short_url_idx ON links (short_url);")

	if err != nil {
		return nil, fmt.Errorf("create index on table links failed: %w", err)
	}

	_, err = db.ExecContext(ctx,
		"GRANT ALL PRIVILEGES ON links TO PUBLIC;")

	if err != nil {
		return nil, fmt.Errorf("grant privilages on table links failed: %w", err)
	}

	return &Links{db: db}, nil
}

func (ls *Links) Begin(ctx context.Context) (*sql.Tx, error) {
	return ls.db.BeginTx(ctx, nil)
}

func (ls *Links) GetLink(ctx context.Context, key string) (string, error) {
	ls.RLock()
	defer ls.RUnlock()

	row := ls.db.QueryRowContext(ctx,
		"SELECT value FROM links WHERE short_url = $1", key)
	var val string

	err := row.Scan(&val)

	if err != nil {
		return "", fmt.Errorf("select link from database failed: %w", err)
	}

	return val, nil
}

func (ls *Links) SetLink(ctx context.Context, key string, link string) error {
	if _, err := ls.GetLink(ctx, key); err != nil {
		ls.Lock()
		defer ls.Unlock()

		_, err := ls.db.ExecContext(ctx,
			"INSERT INTO links (short_url, original_url) VALUES ($1,$2)", key, link)

		if err != nil {
			return fmt.Errorf("insert link to database failed: %w", err)
		}
	}

	return nil
}

func (ls *Links) Count(ctx context.Context) (int, error) {
	//unused
	ls.RLock()
	defer ls.RUnlock()

	row := ls.db.QueryRowContext(ctx,
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

func (ls *Links) CreateShortLinksBatch(ctx context.Context, linksDao dao.LinksDao) (dao.LinksDao, error) {
	ls.Lock()
	defer ls.Unlock()

	tx, err := ls.db.Begin()

	if err != nil {
		return nil, fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO links (short_url, original_url) VALUES ($1,$2)")

	if err != nil {
		return nil, fmt.Errorf("prepare statement failed: %w", err)
	}
	defer stmt.Close()

	for _, linkDao := range linksDao {
		if _, err := url.ParseRequestURI(linkDao.OriginalURL); err != nil {
			return nil, fmt.Errorf("parse request URI failed: %w", err)
		}

		shortUrl := gen.GenKey(linkDao.OriginalURL)
		linkDao.ShortURL = shortUrl
		_, err := stmt.ExecContext(ctx, shortUrl, linkDao.OriginalURL)

		if err != nil {
			return nil, fmt.Errorf("execute statement failed: %w", err)
		}
	}

	tx.Commit()

	return linksDao, nil
}

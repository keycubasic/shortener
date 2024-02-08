package postgre

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/keycubasic/shortener/internal/dao"
	"github.com/keycubasic/shortener/internal/gen"
	"github.com/keycubasic/shortener/internal/storage"
	log "github.com/sirupsen/logrus"
)

type configuration interface {
	GetBaseURL() string
	GetDatabaseDSN() string
}

type Links struct {
	sync.RWMutex
	cfg  configuration
	conn *sql.DB
}

func NewLinks(ctx context.Context, cfg configuration) (*Links, error) {
	log.Infoln("Database DSN:", cfg.GetDatabaseDSN())
	conn, err := sql.Open("pgx", cfg.GetDatabaseDSN())

	if err != nil {
		return nil, fmt.Errorf("database open connection failed: %w", err)
	}

	log.Infoln("Database connected...")

	tx, err := conn.Begin()

	if err != nil {
		return nil, fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS links (
			uuid int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			short_url VARCHAR(8),
			original_url VARCHAR(100)
		);
		`)

	if err != nil {
		return nil, fmt.Errorf("create table links failed: %w", err)
	}

	_, err = tx.ExecContext(ctx, `

		CREATE INDEX IF NOT EXISTS short_url_idx ON links (short_url);
		`)

	if err != nil {
		return nil, fmt.Errorf("create index on field short_url failed: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		CREATE UNIQUE INDEX IF NOT EXISTS original_url_idx ON links (original_url);
		`)

	if err != nil {
		return nil, fmt.Errorf("create unique index on field original_url failed: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		GRANT ALL PRIVILEGES ON links TO PUBLIC;
		`)

	if err != nil {
		return nil, fmt.Errorf("grant privilages on table links failed: %w", err)
	}

	tx.Commit()

	return &Links{cfg: cfg, conn: conn}, nil
}

func (ls *Links) Begin(ctx context.Context) (*sql.Tx, error) {
	return ls.conn.BeginTx(ctx, nil)
}

func (ls *Links) GetLinkUnlocked(ctx context.Context, key string) (string, error) {
	row := ls.conn.QueryRowContext(ctx, `
		SELECT original_url FROM links WHERE short_url = $1
		`, key)
	var val string

	err := row.Scan(&val)

	if err != nil {
		return "", fmt.Errorf("select link from database failed: %w", err)
	}

	return val, nil
}

func (ls *Links) GetLink(ctx context.Context, key string) (string, error) {
	ls.RLock()
	defer ls.RUnlock()

	return ls.GetLinkUnlocked(ctx, key)
}

func (ls *Links) SetLink(ctx context.Context, key string, link string) error {
	ls.Lock()
	defer ls.Unlock()

	if _, err := ls.conn.ExecContext(ctx, `
			INSERT INTO links (short_url, original_url) VALUES ($1,$2)
			`, key, link); err != nil {

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return fmt.Errorf("shortlink already exists: %w", storage.ErrUniqueViolation)
		}

		return fmt.Errorf("insert link to database failed: %w", err)
	}

	return nil
}

func (ls *Links) GetShortLink(ctx context.Context, link string) (string, error) {
	ls.RLock()
	defer ls.RUnlock()

	row := ls.conn.QueryRowContext(ctx, `
		SELECT short_url FROM links WHERE original_url = $1
		`, link)
	var val string

	err := row.Scan(&val)

	if err != nil {
		return "", fmt.Errorf("select shortlink from database failed: %w", err)
	}

	return val, nil
}

func (ls *Links) Count(ctx context.Context) (int, error) {
	//unused (for implementation support)
	ls.RLock()
	defer ls.RUnlock()

	row := ls.conn.QueryRowContext(ctx, `
		SELECT COUNT(1) FROM links
		`)
	var cnt int

	err := row.Scan(&cnt)

	if err != nil {
		return 0, fmt.Errorf("select links count from database failed: %w", err)
	}

	return cnt, nil
}

func (ls *Links) Close() error {
	return ls.conn.Close()
}

func (ls *Links) PingContext(ctx context.Context) error {
	return ls.conn.PingContext(ctx)
}

func (ls *Links) CreateShortLinksBatch(ctx context.Context, linksDao dao.LinksDao) (dao.LinksDao, error) {
	ls.Lock()
	defer ls.Unlock()

	tx, err := ls.conn.Begin()

	if err != nil {
		return nil, fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO links (short_url, original_url) VALUES ($1,$2)
		`)

	if err != nil {
		return nil, fmt.Errorf("prepare statement failed: %w", err)
	}
	defer stmt.Close()

	for _, linkDao := range linksDao {
		if _, err := url.ParseRequestURI(linkDao.OriginalURL); err != nil {
			return nil, fmt.Errorf("parse request URI failed: %w", err)
		}

		shortURL := gen.GenKey(linkDao.OriginalURL)
		linkDao.ShortURL = ls.cfg.GetBaseURL() + "/" + shortURL

		if _, err := ls.GetLinkUnlocked(ctx, shortURL); err != nil {

			if _, err := stmt.ExecContext(ctx, shortURL, linkDao.OriginalURL); err != nil {
				return nil, fmt.Errorf("execute statement failed: %w", err)
			}
		}
	}

	tx.Commit()

	return linksDao, nil
}

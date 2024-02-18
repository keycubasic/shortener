package postgre

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
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
	cfg  configuration
	conn *sqlx.DB
}

func NewLinks(ctx context.Context, cfg configuration) (links *Links, err error) {
	log.Infoln("Database DSN:", cfg.GetDatabaseDSN())
	conn, err := sqlx.Open("pgx", cfg.GetDatabaseDSN())

	if err != nil {
		return nil, fmt.Errorf("postgre storage database open connection: %w", err)
	}

	log.Infoln("Database connected...")

	tx, err := conn.Begin()

	if err != nil {
		return nil, fmt.Errorf("postgre storage begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS links (
			uuid int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			short_url VARCHAR(8),
			original_url VARCHAR(100)
		);
		`)

	if err != nil {
		return nil, fmt.Errorf("postgre storage create table links: %w", err)
	}

	_, err = tx.ExecContext(ctx, `

		CREATE INDEX IF NOT EXISTS short_url_idx ON links (short_url);
		`)

	if err != nil {
		return nil, fmt.Errorf("postgre storage create index on field short_url: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		CREATE UNIQUE INDEX IF NOT EXISTS original_url_idx ON links (original_url);
		`)

	if err != nil {
		return nil, fmt.Errorf("postgre storage create unique index on field original_url: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		GRANT ALL PRIVILEGES ON links TO PUBLIC;
		`)

	if err != nil {
		return nil, fmt.Errorf("postgre storage grant privilages on table links: %w", err)
	}

	err = tx.Commit()

	if err != nil {
		return nil, fmt.Errorf("postgre storage commit: %w", err)
	}

	return &Links{cfg: cfg, conn: conn}, nil
}

func (ls *Links) Begin(ctx context.Context) (*sql.Tx, error) {
	return ls.conn.BeginTx(ctx, nil)
}

func (ls *Links) GetLink(ctx context.Context, key string) (string, error) {
	row := ls.conn.QueryRowContext(ctx, `
		SELECT original_url FROM links WHERE short_url = $1
		`, key)
	var val string

	err := row.Scan(&val)

	if err != nil {
		return "", fmt.Errorf("postgre storage select link from database: %w", err)
	}

	return val, nil
}

func (ls *Links) SetLink(ctx context.Context, key string, link string) error {
	if _, err := ls.conn.ExecContext(ctx, `
			INSERT INTO links (short_url, original_url) VALUES ($1,$2)
			`, key, link); err != nil {

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return fmt.Errorf("postgre storage shortlink already exists: %w", storage.ErrUniqueViolation)
		}

		return fmt.Errorf("postgre storage insert link to database: %w", err)
	}

	return nil
}

func (ls *Links) GetShortLink(ctx context.Context, link string) (string, error) {
	row := ls.conn.QueryRowContext(ctx, `
		SELECT short_url FROM links WHERE original_url = $1
		`, link)
	var val string

	err := row.Scan(&val)

	if err != nil {
		return "", fmt.Errorf("postgre storage select shortlink from database: %w", err)
	}

	return val, nil
}

func (ls *Links) Close() error {
	return ls.conn.Close()
}

func (ls *Links) PingContext(ctx context.Context) error {
	return ls.conn.PingContext(ctx)
}

func (ls *Links) CreateShortLinksBatch(ctx context.Context, linksDao dao.LinksDao) (dao.LinksDao, error) {
	for _, linkDao := range linksDao {
		if _, err := url.ParseRequestURI(linkDao.OriginalURL); err != nil {
			return nil, fmt.Errorf("postgre storage parse request URI: %w", err)
		}

		linkDao.ShortURL = gen.GenKey(linkDao.OriginalURL)
	}

	if _, err := ls.conn.NamedExecContext(ctx, `
	INSERT INTO links (short_url, original_url) VALUES (:short_url, :original_url)
	`, linksDao); err != nil {
		return nil, fmt.Errorf("postgre storage execute statement: %w", err)
	}

	return linksDao, nil
}

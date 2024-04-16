package postgre

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/keycubasic/shortener/internal/dao"
	"github.com/keycubasic/shortener/internal/gen"
	"github.com/keycubasic/shortener/internal/storage"
	"github.com/pressly/goose/v3"
	log "github.com/sirupsen/logrus"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type configuration interface {
	GetBaseURL() string
	GetDatabaseDSN() string
}

type Links struct {
	cfg configuration
	db  *sqlx.DB
}

func NewLinks(ctx context.Context, cfg configuration) (links *Links, err error) {
	log.Info("Database DSN:", cfg.GetDatabaseDSN())
	db, err := sqlx.Open("pgx", cfg.GetDatabaseDSN())
	if err != nil {
		return nil, fmt.Errorf("postgre storage database open connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgre storage database ping connection: %w", err)
	}

	log.Info("Database connected...")
	return &Links{cfg: cfg, db: db}, nil
}

func (ls Links) Migrate(ctx context.Context) error {
	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(log.StandardLogger())

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("postgre storage database migrate set dialect: %w", err)
	}

	if err := goose.UpContext(ctx, ls.db.DB, "migrations"); err != nil {
		return fmt.Errorf("postgre storage database migrate up: %w", err)
	}

	return nil
}

func (ls *Links) GetLink(ctx context.Context, key string) (string, error) {
	row := ls.db.QueryRowContext(ctx, `
		SELECT original_url, is_deleted FROM links WHERE short_url = $1
		`, key)
	var (
		val string
		del sql.NullBool
	)

	err := row.Scan(&val, &del)
	if err != nil {
		return "", fmt.Errorf("postgre storage select link from database: %w", err)
	}

	if del.Valid && del.Bool {
		return "", fmt.Errorf("postgre storage select link from database: %w", storage.ErrGone)
	}

	return val, nil
}

func (ls *Links) SetLink(ctx context.Context, userID int, key, link string) error {
	if _, err := ls.db.ExecContext(ctx, `
		INSERT INTO links (user_id, short_url, original_url) VALUES ($1,$2,$3)
			`, userID, key, link); err != nil {

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return fmt.Errorf("postgre storage shortlink already exists: %w", storage.ErrUniqueViolation)
		}

		return fmt.Errorf("postgre storage insert link to database: %w", err)
	}

	return nil
}

func (ls *Links) GetShortLink(ctx context.Context, link string) (string, error) {
	row := ls.db.QueryRowContext(ctx, `
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
	return ls.db.Close()
}

func (ls *Links) PingContext(ctx context.Context) error {
	return ls.db.PingContext(ctx)
}

func (ls *Links) CreateShortLinksBatch(ctx context.Context, userID int, linksDao dao.LinksDao) (dao.LinksDao, error) {
	for _, linkDao := range linksDao {
		if _, err := url.ParseRequestURI(linkDao.OriginalURL); err != nil {
			return nil, fmt.Errorf("postgre storage parse request URI: %w", err)
		}

		linkDao.UserID = userID
		linkDao.ShortURL = gen.GenKey(linkDao.OriginalURL)

	}

	if _, err := ls.db.NamedExecContext(ctx, `
		INSERT INTO links (user_id, short_url, original_url) VALUES (:user_id, :short_url, :original_url)
	`, linksDao); err != nil {
		return nil, fmt.Errorf("postgre storage insert into links batch: %w", err)
	}

	return linksDao, nil
}

func (ls *Links) AddUser(ctx context.Context) (id int, err error) {
	err = ls.db.QueryRowContext(ctx, `
		INSERT INTO users DEFAULT VALUES RETURNING user_id
	`).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("postgre storage insert into users: %w", err)
	}
	return
}

func (ls *Links) GetUserURLs(ctx context.Context, userID int) (linksDao dao.LinksDaoBatch, err error) {
	rows, err := ls.db.QueryContext(ctx, `
		SELECT short_url, original_url FROM links WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("postgre storage select user URLs from database: %w", err)
	}

	var shortURL, originalURL string

	for rows.Next() {
		err := rows.Scan(&shortURL, &originalURL)
		if err != nil {
			return nil, fmt.Errorf("postgre storage get user URLs scan row: %w", err)
		}

		linksDao = append(linksDao, &dao.LinkDaoBatch{ShortURL: shortURL, OriginalURL: originalURL})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgre storage get user URLs after scan rows: %w", err)
	}

	return
}

func (ls *Links) DeleteShortLinksBatch(ctx context.Context, linksDao dao.LinksDao) error {
	var values []string
	var args []any
	for i, linkDao := range linksDao {
		base := i * 3
		params := fmt.Sprintf("($%d, $%d, $%d)", base+1, base+2, base+3)
		values = append(values, params)
		args = append(args, fmt.Sprint(linkDao.UserID), linkDao.ShortURL, fmt.Sprint(linkDao.DeletedFlag))
	}

	query := `
		UPDATE links AS ls SET is_deleted = CAST(v.is_deleted AS BOOLEAN)
		  FROM (VALUES ` + strings.Join(values, ",\n") + `)
		    AS v (user_id, short_url, is_deleted)
		 WHERE ls.user_id = CAST(v.user_id AS INTEGER)
           AND ls.short_url = v.short_url`

	_, err := ls.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("postgre storage delete from links batch: %w", err)
	}

	return nil
}

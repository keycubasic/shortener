// Package app contains the basic functions for the application.
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/dao"
	"github.com/keycubasic/shortener/internal/dto"
	"github.com/keycubasic/shortener/internal/gen"
	"github.com/keycubasic/shortener/internal/mapper"
	log "github.com/sirupsen/logrus"
)

// Interface of database
//
//go:generate mockgen -source app.go -destination mock_storage.go -package mock Database
type Database interface {
	// Migrate migrates database objects.
	Migrate(ctx context.Context) error
	// PingContext checks the database connection.
	PingContext(ctx context.Context) error
	// GetLink returns the original shortened link from the database.
	GetLink(ctx context.Context, key string) (string, error)
	// SetLink sets a new link-shortening pair to the database.
	SetLink(ctx context.Context, userID int, key, link string) error
	// GetShortLink returns the shortening of the original link from the database.
	GetShortLink(ctx context.Context, link string) (string, error)
	// CreateShortLinksBatch creates link-shortening pairs in the database in batches.
	CreateShortLinksBatch(ctx context.Context, userID int, linksDao dao.LinksDao) (dao.LinksDao, error)
	// AddUser adds a new user to the database.
	AddUser(ctx context.Context) (int, error)
	// GetUserURLs returns user URLs from the database.
	GetUserURLs(ctx context.Context, userID int) (dao.LinksDaoBatch, error)
	// DeleteShortLinksBatch sets the deletion flag on links in batch mode.
	DeleteShortLinksBatch(ctx context.Context, linksDao dao.LinksDao) error
	// Close closes the connection to the database.
	Close() error
}

// App stores an interface for application configuration and data storage,
// as well as a channel for working with the link removal routine.
type App struct {
	cfg    *config.Cfg
	db     Database
	linkCh chan dao.LinkDao
}

// NewApp creates a new App object
func NewApp(ctx context.Context, cfg *config.Cfg, db Database) *App {
	bufferSize := cfg.BufferSize

	instance := &App{
		cfg: cfg,
		db:  db,
	}

	if bufferSize > 0 {
		instance.linkCh = make(chan dao.LinkDao, bufferSize)
		go instance.batchDelete(ctx)
	}

	return instance
}

// Ping checks the availability of the data store.
func (a *App) Ping(ctx context.Context) error {
	return a.db.PingContext(ctx)
}

// GetLink returns the original link by its shortening.
func (a *App) GetLink(ctx context.Context, id string) (string, error) {
	link, err := a.db.GetLink(ctx, id)
	if err != nil {
		return "", fmt.Errorf("app get link: %w", err)
	}

	return link, nil
}

// CreateShortLink сreates a shortening for a link.
func (a *App) CreateShortLink(ctx context.Context, userID int, link string) (string, error) {
	key := gen.GenKey(link)

	if err := a.db.SetLink(ctx, userID, key, link); err != nil {
		return "", fmt.Errorf("app create short link: %w", err)
	}

	return key, nil
}

// GetShortLink returns the shortening of the original link.
func (a *App) GetShortLink(ctx context.Context, link string) (string, error) {
	shortLink, err := a.db.GetShortLink(ctx, link)
	if err != nil {
		return "", fmt.Errorf("app get short link: %w", err)
	}

	return shortLink, nil
}

// CreateShortLinksBatch creates shortenings for links in batch mode.
func (a *App) CreateShortLinksBatch(ctx context.Context, userID int, linksDto dto.LinksDto) (dto.LinksDto, error) {
	capacity := len(linksDto)
	reqLinksDao := make(dao.LinksDao, 0, capacity)
	resLinksDto := make(dto.LinksDto, 0, capacity)

	for _, linkDto := range linksDto {
		linkDao := mapper.LinkDtoToLinkDao(linkDto)
		reqLinksDao = append(reqLinksDao, linkDao)
	}

	resLinksDao, err := a.db.CreateShortLinksBatch(ctx, userID, reqLinksDao)
	if err != nil {
		return nil, fmt.Errorf("app create short links batch: %w", err)
	}

	for _, linkDao := range resLinksDao {
		linkDto := mapper.LinkDaoToLinkDto(*linkDao, a.cfg.BaseURL)
		resLinksDto = append(resLinksDto, *linkDto)
	}

	return resLinksDto, nil
}

// AddUser adds a new user.
func (a *App) AddUser(ctx context.Context) (int, error) {
	id, err := a.db.AddUser(ctx)
	if err != nil {
		return 0, fmt.Errorf("app add user: %w", err)
	}

	return id, nil
}

// GetUserURLs returns user URLs.
func (a *App) GetUserURLs(ctx context.Context, userID int) (dto.LinksDto, error) {
	linksDao, err := a.db.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("app get user URLs: %w", err)
	}

	linksDto := make([]dto.LinkDto, 0, len(linksDao))

	for _, linkDao := range linksDao {
		linkDto := mapper.LinkDaoToLinkDtoURLs(*linkDao, a.cfg.BaseURL)
		linksDto = append(linksDto, *linkDto)
	}

	return linksDto, nil
}

// AppendChan sends links marked for deletion to the channel for further processing.
func (a *App) AppendChan(ctx context.Context, userID int, slice dto.Slice) {
	for _, s := range slice {
		linkDao := mapper.LinkToLinkDao(userID, s)
		if a.linkCh != nil {
			a.linkCh <- *linkDao
		}
	}
}

func (a *App) deleteShortLinksBatch(ctx context.Context, linksDao dao.LinksDao) {
	if len(linksDao) > 0 {
		a.db.DeleteShortLinksBatch(ctx, linksDao)
		log.Infof("Goroutine batchDelete processes %d links", len(linksDao))
	}
}

func (a *App) batchDelete(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(a.cfg.TickerDuration) * time.Second)
	linksDao := make(dao.LinksDao, 0, a.cfg.BufferSize)

	for {
		select {
		case <-ctx.Done():
			ctxClosing, closingCancel := context.
				WithTimeout(context.Background(), time.Duration(a.cfg.ClosingTimeout)*time.Second)
			defer closingCancel()

			if err := a.Ping(ctxClosing); err == nil {
				a.deleteShortLinksBatch(ctxClosing, linksDao)
			}

			log.Info("Goroutine batchDelete stops")
			return
		case linkDao := <-a.linkCh:
			linksDao = append(linksDao, &linkDao)
		case <-ticker.C:
			a.deleteShortLinksBatch(ctx, linksDao)
			linksDao = nil
		}
	}
}

func (a *App) Close() error {
	return a.db.Close()
}

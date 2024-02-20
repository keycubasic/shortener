package app

import (
	"context"
	"fmt"

	"github.com/keycubasic/shortener/internal/dao"
	"github.com/keycubasic/shortener/internal/dto"
	"github.com/keycubasic/shortener/internal/gen"
	"github.com/keycubasic/shortener/internal/mapper"
)

type configuration interface {
	GetBaseURL() string
}

//go:generate mockgen -source app.go -destination mock_storage.go -package mock database
type Database interface {
	GetLink(ctx context.Context, key string) (string, error)
	SetLink(ctx context.Context, userID int, key, link string) error
	GetShortLink(ctx context.Context, link string) (string, error)
	Close() error
	PingContext(ctx context.Context) error
	CreateShortLinksBatch(ctx context.Context, userID int, linksDao dao.LinksDao) (dao.LinksDao, error)
	AddUser(ctx context.Context) (int, error)
	GetUserURLs(ctx context.Context, userID int) (dao.LinksDao, error)
}

type App struct {
	cfg configuration
	db  Database
}

func NewApp(cfg configuration, db Database) *App {
	return &App{
		cfg: cfg,
		db:  db,
	}
}

func (a *App) Ping(ctx context.Context) error {
	return a.db.PingContext(ctx)
}

func (a *App) GetLink(ctx context.Context, id string) (string, error) {
	link, err := a.db.GetLink(ctx, id)

	if err != nil {
		return "", fmt.Errorf("app get link: %w", err)
	}

	return link, nil
}

func (a *App) CreateShortLink(ctx context.Context, userID int, link string) (string, error) {
	key := gen.GenKey(link)

	if err := a.db.SetLink(ctx, userID, key, link); err != nil {
		return "", fmt.Errorf("app create short link: %w", err)
	}

	return key, nil
}

func (a *App) GetShortLink(ctx context.Context, link string) (string, error) {
	shortLink, err := a.db.GetShortLink(ctx, link)

	if err != nil {
		return "", fmt.Errorf("app get short link: %w", err)
	}

	return shortLink, nil
}

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
		linkDto := mapper.LinkDaoToLinkDto(*linkDao, a.cfg.GetBaseURL())
		resLinksDto = append(resLinksDto, *linkDto)
	}

	return resLinksDto, nil
}

func (a *App) AddUser(ctx context.Context) (int, error) {
	id, err := a.db.AddUser(ctx)

	if err != nil {
		return 0, fmt.Errorf("app add user: %w", err)
	}

	return id, nil
}

func (a *App) GetUserURLs(ctx context.Context, userID int) (dto.LinksDto, error) {
	linksDao, err := a.db.GetUserURLs(ctx, userID)

	if err != nil {
		return nil, fmt.Errorf("app get user URLs: %w", err)
	}

	linksDto := make([]dto.LinkDto, 0, len(linksDao))

	for _, linkDao := range linksDao {
		linkDto := mapper.LinkDaoToLinkDtoURLs(*linkDao, a.cfg.GetBaseURL())
		linksDto = append(linksDto, *linkDto)
	}

	return linksDto, nil
}

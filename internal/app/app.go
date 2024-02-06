package app

import (
	"context"
	"fmt"

	"github.com/keycubasic/shortener/internal/dao"
	"github.com/keycubasic/shortener/internal/dto"
	"github.com/keycubasic/shortener/internal/gen"
	"github.com/keycubasic/shortener/internal/mapper"
)

//go:generate mockgen -source app.go -destination mock_storage.go -package mock database
type DB interface {
	GetLink(ctx context.Context, key string) (string, error)
	SetLink(ctx context.Context, key string, link string) error
	Close() error
	PingContext(ctx context.Context) error
	CreateShortLinksBatch(ctx context.Context, linksDao dao.LinksDao) (dao.LinksDao, error)
}

type App struct {
	db DB
}

func NewApp(db DB) *App {
	return &App{
		db: db,
	}
}

func (a *App) Ping(ctx context.Context) error {
	return a.db.PingContext(ctx)
}

func (a *App) GetLink(ctx context.Context, id string) (string, error) {
	return a.db.GetLink(ctx, id)
}

func (a *App) CreateShortLink(ctx context.Context, link string) (string, error) {
	key := gen.GenKey(link)
	a.db.SetLink(ctx, key, link)
	return key, nil
}

func (a *App) CreateShortLinksBatch(ctx context.Context, linksDto dto.LinksDto) (dto.LinksDto, error) {
	reqLinksDao := dao.NewLinksDao()
	resLinksDto := dto.NewLinksDto()

	for _, linkDto := range linksDto {
		linkDao := mapper.LinkDtoToLinkDao(linkDto)
		reqLinksDao = append(reqLinksDao, linkDao)
	}

	resLinksDao, err := a.db.CreateShortLinksBatch(ctx, reqLinksDao)

	if err != nil {
		return nil, fmt.Errorf("create short links batch fail: %w", err)
	}

	for _, linkDao := range resLinksDao {
		linkDto := mapper.LinkDaoToLinkDto(*linkDao)
		resLinksDto = append(resLinksDto, *linkDto)
	}

	return resLinksDto, nil
}

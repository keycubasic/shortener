package storage

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/keycubasic/shortener/internal/dao"
	"github.com/keycubasic/shortener/internal/gen"
	"github.com/keycubasic/shortener/internal/loader"
	log "github.com/sirupsen/logrus"
)

var ErrUniqueViolation = errors.New("unique violation")

type Cfg interface {
	GetBaseURL() string
	GetFilePath() string
}

type Links struct {
	sync.RWMutex
	cfg      Cfg
	links    map[string]string
	producer *loader.Producer
}

func NewLinks(ctx context.Context, cfg Cfg) (*Links, error) {
	ls := &Links{
		cfg:   cfg,
		links: make(map[string]string),
	}

	if cfg.GetFilePath() != "" {
		consumer, err := loader.NewConsumer(ls, cfg.GetFilePath())

		if err != nil {
			return nil, fmt.Errorf("storage failed to create consumer: %w", err)
		}

		if err := consumer.Load(ctx); err != nil {
			return nil, fmt.Errorf("storage failed to load data from file: %w", err)
		}

		if err := consumer.Close(); err != nil {
			return nil, fmt.Errorf("storage failed to close consumer: %w", err)
		}

		log.Infoln("Records count preloaded from disk storage:", ls.Count(ctx))

		// Init producer
		if ls.producer, err = loader.NewProducer(cfg.GetFilePath()); err != nil {
			return nil, fmt.Errorf("storage failed to create producer: %w", err)
		}
	}

	log.Infoln("Internal storage created...")

	return ls, nil
}

func (ls *Links) GetLink(ctx context.Context, key string) (string, error) {
	ls.RLock()
	defer ls.RUnlock()

	val, ok := ls.links[key]

	if !ok {
		return "", errors.New("get link from map failed")
	}

	return val, nil
}

func (ls *Links) SetLink(ctx context.Context, key string, link string) error {
	ls.Lock()
	defer ls.Unlock()

	if _, err := ls.GetShortLink(ctx, link); err == nil {
		return fmt.Errorf("shortlink already exists: %w", ErrUniqueViolation)
	}

	ls.links[key] = link

	if ls.producer != nil {
		id := ls.Count(ctx) + 1

		if err := ls.producer.Save(ctx, id, key, link); err != nil {
			return fmt.Errorf("storage failed to save data into file: %w", err)
		}
	}

	return nil
}

func (ls *Links) GetShortLink(ctx context.Context, link string) (string, error) {
	for k, v := range ls.links {
		if v == link {
			return k, nil
		}
	}
	return "", fmt.Errorf("shortlink not found")
}

func (ls *Links) Count(ctx context.Context) int {
	return len(ls.links)
}

func (ls *Links) Close() error {
	// for implementation support
	return nil
}

func (ls *Links) PingContext(ctx context.Context) error {
	// for implementation support
	return nil
}

func (ls *Links) CreateShortLinksBatch(ctx context.Context, linksDao dao.LinksDao) (dao.LinksDao, error) {
	ls.Lock()
	defer ls.Unlock()

	for _, linkDao := range linksDao {
		if _, err := url.ParseRequestURI(linkDao.OriginalURL); err != nil {
			return nil, fmt.Errorf("parse request URI failed: %w", err)
		}

		shortURL := gen.GenKey(linkDao.OriginalURL)
		linkDao.ShortURL = ls.cfg.GetBaseURL() + "/" + shortURL
		if err := ls.SetLink(ctx, shortURL, linkDao.OriginalURL); err != nil {
			return nil, fmt.Errorf("set link failed: %w", err)
		}
	}
	return linksDao, nil
}

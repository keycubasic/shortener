// Package storage contains functions for working with a hash table.
package storage

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/keycubasic/shortener/internal/dao"
	"github.com/keycubasic/shortener/internal/gen"
	"github.com/keycubasic/shortener/internal/loader"
	log "github.com/sirupsen/logrus"
)

// Variables
var (
	// ErrUniqueViolation storage unique violation error.
	ErrUniqueViolation = errors.New("storage unique violation")
	// ErrGone link already gone error.
	ErrGone = errors.New("link already gone")
)

// Interface of configuration
type configuration interface {
	// GetBaseURL returns the base address of the redirection service.
	GetBaseURL() string
	// GetFilePath returns the path to the storage file.
	GetFilePath() string
}

// Links stores the mutex, application configuration, hash table and file storage producer.
type Links struct {
	sync.RWMutex
	cfg      configuration
	links    map[string]string
	producer *loader.Producer
}

// NewLinks creates a new Links object
func NewLinks(ctx context.Context, cfg configuration) (*Links, error) {
	ls := &Links{
		cfg:   cfg,
		links: make(map[string]string),
	}

	if cfg.GetFilePath() != "" {
		consumer, err := loader.NewConsumer(ls, cfg.GetFilePath())
		if err != nil {
			return nil, fmt.Errorf("storage create consumer: %w", err)
		}

		if err = consumer.Load(ctx); err != nil {
			return nil, fmt.Errorf("storage load data from file: %w", err)
		}

		if err = consumer.Close(); err != nil {
			return nil, fmt.Errorf("storage close consumer: %w", err)
		}

		log.Info("Records count preloaded from disk storage:", ls.count())

		// Init producer
		if ls.producer, err = loader.NewProducer(cfg.GetFilePath()); err != nil {
			return nil, fmt.Errorf("storage create producer: %w", err)
		}
	}

	log.Info("Internal storage created...")

	return ls, nil
}

// GetLink returns the original link by its shortening from the hash table.
func (ls *Links) GetLink(_ context.Context, key string) (string, error) {
	ls.RLock()
	defer ls.RUnlock()

	val, ok := ls.links[key]

	if !ok {
		return "", fmt.Errorf("storage get link from map")
	}

	return val, nil
}

// SetLink sets a new link-shortenening pair to the hash table.
func (ls *Links) SetLink(ctx context.Context, userID int, key string, link string) error {
	ls.Lock()
	defer ls.Unlock()

	if _, err := ls.GetShortLink(ctx, link); err == nil {
		return fmt.Errorf("storage shortlink already exists: %w", ErrUniqueViolation)
	}

	ls.links[key] = link

	if ls.producer != nil {
		id := ls.count() + 1

		if err := ls.producer.Save(ctx, id, key, link); err != nil {
			return fmt.Errorf("storage save data into file: %w", err)
		}
	}

	return nil
}

// GetShortLink returns the shortening of the original link from the hash table.
func (ls *Links) GetShortLink(_ context.Context, link string) (string, error) {
	for k, v := range ls.links {
		if v == link {
			return k, nil
		}
	}
	return "", fmt.Errorf("storage shortlink not found")
}

// count
func (ls *Links) count() int {
	return len(ls.links)
}

// Close
func (ls *Links) Close() error {
	// for implementation support
	return nil
}

// PingContext
func (ls *Links) PingContext(_ context.Context) error {
	// for implementation support
	return nil
}

// CreateShortLinksBatch creates link-shortening pairs in a hash table in batch mode.
func (ls *Links) CreateShortLinksBatch(ctx context.Context, userID int, linksDao dao.LinksDao) (dao.LinksDao, error) {
	for _, linkDao := range linksDao {
		if _, err := url.ParseRequestURI(linkDao.OriginalURL); err != nil {
			return nil, fmt.Errorf("storage parse request URI: %w", err)
		}

		shortURL := gen.GenKey(linkDao.OriginalURL)
		linkDao.ShortURL = strings.Join([]string{ls.cfg.GetBaseURL(), shortURL}, "/")

		if err := ls.SetLink(ctx, 0, shortURL, linkDao.OriginalURL); err != nil {
			return nil, fmt.Errorf("storage set link: %w", err)
		}
	}
	return linksDao, nil
}

// AddUser
func (ls *Links) AddUser(_ context.Context) (int, error) {
	// for implementation support
	return 1, nil
}

// GetUserURLs
func (ls *Links) GetUserURLs(_ context.Context, _ int) (dao.LinksDaoBatch, error) {
	// for implementation support
	linksDao := make([]*dao.LinkDaoBatch, 0)
	return linksDao, nil
}

// DeleteShortLinksBatch
func (ls *Links) DeleteShortLinksBatch(_ context.Context, _ dao.LinksDao) error {
	// for implementation support
	return nil
}

// Migrate
func (ls *Links) Migrate(_ context.Context) error {
	// for implementation support
	return nil
}

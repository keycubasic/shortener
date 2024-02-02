package storage_legacy

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/keycubasic/shortener/internal/loader"
	log "github.com/sirupsen/logrus"
)

type Cfg interface {
	GetFilePath() string
}

type Links struct {
	sync.RWMutex
	links    map[string]string
	producer *loader.Producer
}

func NewLinks(cfg Cfg) (*Links, error) {
	ls := &Links{links: make(map[string]string)}

	if cfg.GetFilePath() != "" {
		consumer, err := loader.NewConsumer(ls, cfg.GetFilePath())

		if err != nil {
			return nil, fmt.Errorf("storage failed to create consumer: %w", err)
		}

		if err := consumer.Load(); err != nil {
			return nil, fmt.Errorf("storage failed to load data from file: %w", err)
		}

		if err := consumer.Close(); err != nil {
			return nil, fmt.Errorf("storage failed to close consumer: %w", err)
		}

		log.Infoln("Records count preloaded from disk storage:", ls.Count())

		// Init producer
		if ls.producer, err = loader.NewProducer(cfg.GetFilePath()); err != nil {
			return nil, fmt.Errorf("storage failed to create producer: %w", err)
		}
	}

	return ls, nil
}

func (ls *Links) GetLink(key string) (string, error) {
	ls.RLock()
	defer ls.RUnlock()
	val, ok := ls.links[key]
	if !ok {
		return "", errors.New("get link from map failed")
	}
	return val, nil
}

func (ls *Links) SetLink(key string, link string) error {
	ls.Lock()
	defer ls.Unlock()
	ls.links[key] = link

	if ls.producer != nil {
		id := ls.Count() + 1

		if err := ls.producer.Save(id, key, link); err != nil {
			return fmt.Errorf("storage failed to save data into file: %w", err)
		}
	}

	return nil
}

func (ls *Links) Count() int {
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

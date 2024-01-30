package storage

import (
	"fmt"
	"sync"

	"github.com/keycubasic/shortener/internal/loader"
	log "github.com/sirupsen/logrus"
)

type configuration interface {
	GetFilePath() string
}

type Links struct {
	sync.RWMutex
	links    map[string]string
	producer *loader.Producer
}

func NewLinks(cfg configuration) (*Links, error) {
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

func (ls *Links) GetLink(key string) (string, bool) {
	ls.RLock()
	defer ls.RUnlock()
	val, ok := ls.links[key]
	return val, ok
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

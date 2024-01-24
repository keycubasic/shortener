package storage

import (
	"sync"

	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/loader"
	log "github.com/sirupsen/logrus"
)

type Links struct {
	sync.RWMutex
	links map[string]string
}

func NewLinks() *Links {
	return &Links{links: make(map[string]string)}
}

func (ls *Links) GetLink(key string) (string, bool) {
	ls.RLock()
	defer ls.RUnlock()
	val, ok := ls.links[key]
	return val, ok
}

func (ls *Links) SetLink(key string, link string) {
	ls.Lock()
	defer ls.Unlock()
	ls.links[key] = link
}

func (ls *Links) Count() int {
	return len(ls.links)
}

func (ls *Links) Load(cfg *config.Cfg) {
	if cfg.GetFilePath() != "" {
		consumer, err := loader.NewConsumer(cfg.GetFilePath())

		if err != nil {
			log.Warnln(err.Error())
		}

		if err := consumer.Load(ls); err != nil {
			log.Warnln(err.Error())
		}

		if err := consumer.Close(); err != nil {
			log.Warnln(err.Error())
		}

		log.Infoln("Records count preloaded from disk storage:", ls.Count())
	}
}

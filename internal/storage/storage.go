package storage

import "sync"

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

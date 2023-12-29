package storage

import "sync"

type Links struct {
	sync.Mutex
	l map[string]string
}

func NewLinks() (*Links, error) {
	sl := &Links{l: make(map[string]string)}
	return sl, nil
}

func (sl *Links) GetLink(key string) (string, bool) {
	defer sl.Unlock()
	sl.Lock()
	val, ok := sl.l[key]
	return val, ok
}

func (sl *Links) SetLink(key string, link string) {
	defer sl.Unlock()
	sl.Lock()
	sl.l[key] = link
}

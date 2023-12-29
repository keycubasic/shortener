package main

import (
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/server"
	"github.com/keycubasic/shortener/internal/storage"
)

func main() {
	cfg, _ := config.NewCfg()
	cfg.ParseCfg()
	db, _ := storage.NewLinks()
	server.StartServer(cfg, db)
}

package main

import (
	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/handler"
	"github.com/keycubasic/shortener/internal/storage"
)

func main() {
	cfg := config.NewCfg()
	cfg.ParseCfg()
	db := storage.NewLinks()
	db.Load(cfg)
	app := app.NewApp(db)
	handler.RunServer(app, cfg)
}

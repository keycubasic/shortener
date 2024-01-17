package main

import (
	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/handler"
	"github.com/keycubasic/shortener/internal/storage"
)

func main() {
	db := storage.NewLinks()
	app := app.NewApp(db)
	cfg := config.NewCfg()
	cfg.ParseCfg()
	handler.RunServer(app, cfg)
}

package main

import (
	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/handler"
	"github.com/keycubasic/shortener/internal/storage"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := config.NewCfg()
	cfg.ParseCfg()
	db, err := storage.NewLinks(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	app := app.NewApp(db)
	handler.RunServer(app, cfg)
}

package main

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/handler"
	"github.com/keycubasic/shortener/internal/storage"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := config.NewCfg()
	cfg.ParseCfg()

	dbLegacy, err := storage.NewLinks(cfg)

	if err != nil {
		log.Fatal(err.Error())
	}

	app := app.NewApp(cfg, dbLegacy)
	handler.RunServer(app, cfg)
}

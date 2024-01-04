package main

import (
	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/handler"
)

func main() {
	cfg := config.NewCfg()
	cfg.ParseCfg()
	db := app.ConnectStorage()
	handler.RunServer(cfg, db)
}

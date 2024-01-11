package main

import (
	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/handler"
)

func main() {
	app := app.NewApp()
	cfg := config.NewCfg()
	cfg.ParseCfg()
	handler.RunServer(app, cfg)
}

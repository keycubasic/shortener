package main

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/handler"
	"github.com/keycubasic/shortener/internal/storage"
	"github.com/keycubasic/shortener/internal/storage/postgre"
	log "github.com/sirupsen/logrus"
)

func connectDB(ctx context.Context, cfg *config.Cfg) (app.Connection, error) {
	if cfg.GetDatabaseDSN() != "" {
		return postgre.NewLinks(ctx, cfg)
	} else {
		return storage.NewLinks(ctx, cfg)
	}
}

func main() {
	ctx := context.Background()
	cfg := config.NewCfg()
	cfg.ParseCfg()
	conn, err := connectDB(ctx, cfg)

	if err != nil {
		log.Fatal(err.Error())
	}
	defer conn.Close()

	app := app.NewApp(conn)
	handler.RunServer(ctx, app, cfg)
}

package main

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/handler"
	"github.com/keycubasic/shortener/internal/storage"
	"github.com/keycubasic/shortener/internal/storageinternal"
	log "github.com/sirupsen/logrus"
)

func connectDB(ctx context.Context, cfg *config.Cfg) (app.DB, error) {
	if cfg.GetDatabaseDSN() != "" {
		return storage.NewLinks(ctx, cfg)
	} else {
		return storageinternal.NewLinks(ctx, cfg)
	}
}

func main() {
	ctx := context.Background()
	cfg := config.NewCfg()
	cfg.ParseCfg()
	db, err := connectDB(ctx, cfg)

	if err != nil {
		log.Fatal(err.Error())
	}

	defer db.Close()

	app := app.NewApp(db)
	handler.RunServer(ctx, app, cfg)
}

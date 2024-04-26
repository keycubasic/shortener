// Main package of application
package main

import (
	"context"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/handler"
	"github.com/keycubasic/shortener/internal/storage"
	"github.com/keycubasic/shortener/internal/storage/postgre"
	log "github.com/sirupsen/logrus"
)

func connectDB(ctx context.Context, cfg *config.Cfg) (app.Database, error) {
	if cfg.GetDatabaseDSN() != "" {
		return postgre.NewLinks(ctx, cfg)
	} else {
		return storage.NewLinks(ctx, cfg)
	}
}

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM)
	defer stop()

	cfg := config.NewCfg()
	cfg.ParseCfg()
	db, err := connectDB(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	err = db.Migrate(ctx)
	if err != nil {
		log.Fatal(err)
	}

	app := app.NewApp(ctx, cfg, db)
	handler.RunServer(ctx, cfg, app)
}

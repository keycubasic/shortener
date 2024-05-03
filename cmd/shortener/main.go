// Main package of application
package main

import (
	"context"
	"fmt"
	"os/signal"
	"runtime/debug"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/handler"
	"github.com/keycubasic/shortener/internal/storage"
	"github.com/keycubasic/shortener/internal/storage/postgre"
	log "github.com/sirupsen/logrus"
)

var (
	buildVersion = "N/A"
	buildDate    = readBuildInfo("vcs.time")
	// buildCommit         = readBuildInfo("vcs.revision")
	// go build -ldflags="-X main.buildCommit=$(git rev-parse HEAD)"
	buildCommit = "N/A"
)

func readBuildInfo(key string) string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == key {
				return setting.Value
			}
		}
	}
	return "N/A"
}

func init() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

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

	cfg := new(config.Cfg)
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

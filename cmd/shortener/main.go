package main

import (
	"database/sql"

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

	log.Infoln("Database DSN:", cfg.GetDatabaseDSN())
	db, err := sql.Open("pgx", cfg.GetDatabaseDSN())

	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Database connection failed")
	}
	log.Infoln("Database connected")

	dbLegacy, err := storage.NewLinks(cfg)

	if err != nil {
		log.Fatal(err.Error())
	}

	app := app.NewApp(dbLegacy, db)
	handler.RunServer(app, cfg)
}

package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/handler"
	"github.com/keycubasic/shortener/internal/storage"
)

func ConfigureRouter(cfg *config.Cfg, db *storage.Links) http.Handler {
	linkHandler := handler.URLHandler{Cfg: cfg, DB: db}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/{id}", linkHandler.GetOriginalURL)
	r.Post("/", linkHandler.ShortenURL)
	return r
}

func StartServer(cfg *config.Cfg, db *storage.Links) {
	r := ConfigureRouter(cfg, db)
	log.Printf("Server listening address %s with base URL %s", cfg.GetAddress(), cfg.GetBaseUrl())
	if err := http.ListenAndServe(cfg.GetAddress(), r); err != nil {
		log.Fatal(err)
	}
	log.Println("Server stops")
}

package handler

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/storage"
)

type URLHandler struct {
	DB  *storage.Links
	Cfg *config.Cfg
}

func NewURLHandler(cfg *config.Cfg, db *storage.Links) *URLHandler {
	return &URLHandler{Cfg: cfg, DB: db}
}

func (h *URLHandler) GetOriginalURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	link, ok := app.GetLink(h.DB, id)
	if !ok {
		http.Error(w, "target link not found", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", link)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *URLHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	link := string(body)
	if link == "" {
		http.Error(w, "body should't be empty", http.StatusBadRequest)
		return
	}
	if strings.Contains(link, "\n") {
		http.Error(w, "body contains too many lines", http.StatusBadRequest)
		return
	}
	if _, err := url.ParseRequestURI(link); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	key, err := app.CreateShortLink(h.DB, link)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // http.StatusInternalServerError
		return
	}
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	resBody := h.Cfg.GetBaseURL() + "/" + key
	if _, err := w.Write([]byte(resBody)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func getRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	return r
}

func RunServer(cfg *config.Cfg, db *storage.Links) {
	h := NewURLHandler(cfg, db)
	r := getRouter()
	r.Get("/{id}", h.GetOriginalURL)
	r.Post("/", h.ShortenURL)
	log.Printf("Server listening address %s with base URL %s", cfg.GetAddress(), cfg.GetBaseURL())
	if err := http.ListenAndServe(cfg.GetAddress(), r); err != nil {
		log.Fatal(err)
	}
	log.Println("Server stops")
}

package handler

import (
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	_ "github.com/go-chi/chi/middleware"
	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/dto"
	"github.com/keycubasic/shortener/internal/middleware"
)

type URLHandler struct {
	app *app.App
	cfg *config.Cfg
}

func NewURLHandler(app *app.App, cfg *config.Cfg) *URLHandler {
	return &URLHandler{app: app, cfg: cfg}
}

func (h *URLHandler) GetOriginalURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	link, ok := h.app.GetLink(id)

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

	if _, err := url.ParseRequestURI(link); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	key, err := h.app.CreateShortLink(link)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	resBody := h.cfg.GetBaseURL() + "/" + key

	if _, err := w.Write([]byte(resBody)); err != nil {
		http.Error(w, "unexpected error occurred. try again", http.StatusBadRequest)
		return
	}
}

func (h *URLHandler) ShortenURLJson(w http.ResponseWriter, r *http.Request) {
	req := dto.NewRequest()

	if err := req.Unmarshal(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if ok, err := req.ValidateRequest(); !ok {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	key, err := h.app.CreateShortLink(req.URL)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := dto.NewResponse(h.cfg.GetBaseURL() + "/" + key)
	resBody, err := res.Marshal()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write([]byte(resBody)); err != nil {
		http.Error(w, "unexpected error occurred. try again", http.StatusBadRequest)
		return
	}
}

func (h *URLHandler) getRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/{id}", h.GetOriginalURL)
	r.Post("/", h.ShortenURL)
	r.Post("/api/shorten", h.ShortenURLJson)
	return r
}

func RunServer(app *app.App, cfg *config.Cfg) {
	h := NewURLHandler(app, cfg)
	r := h.getRouter()
	log.Printf("Server listening address %s with base URL %s", cfg.GetAddress(), cfg.GetBaseURL())

	if err := http.ListenAndServe(cfg.GetAddress(), r); err != nil {
		log.Fatal(err)
	}

	log.Println("Server stops")
}

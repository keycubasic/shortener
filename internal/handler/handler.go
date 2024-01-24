package handler

import (
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	_ "github.com/go-chi/chi/middleware"
	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/dto"
	"github.com/keycubasic/shortener/internal/loader"
	"github.com/keycubasic/shortener/internal/middleware"
	log "github.com/sirupsen/logrus"
)

type URLHandler struct {
	app      *app.App
	cfg      *config.Cfg
	producer *loader.Producer
}

func newURLHandler(app *app.App, cfg *config.Cfg) *URLHandler {
	return &URLHandler{app: app, cfg: cfg}
}

func (h *URLHandler) initProducer() error {

	if h.cfg.GetFilePath() != "" {
		var err error

		if h.producer, err = loader.NewProducer(h.cfg.GetFilePath()); err != nil {
			return err
		}
	}

	return nil
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

	if h.producer != nil {
		id := h.app.Count() + 1
		h.producer.Save(id, key, req.URL)
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
	r.Use(middleware.Gzip)
	r.Get("/{id}", h.GetOriginalURL)
	r.Post("/", h.ShortenURL)
	r.Post("/api/shorten", h.ShortenURLJson)
	return r
}

func RunServer(app *app.App, cfg *config.Cfg) {
	h := newURLHandler(app, cfg)

	if err := h.initProducer(); err != nil {
		log.Fatal(err)
	}

	r := h.getRouter()
	log.Infof("Server listening address %s with base URL %s and disk store support at %s", cfg.GetAddress(), cfg.GetBaseURL(), cfg.GetFilePath())

	if err := http.ListenAndServe(cfg.GetAddress(), r); err != nil {
		log.Fatal(err)
	}

	log.Println("Server stops")
}

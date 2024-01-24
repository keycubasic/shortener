package handler

import (
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	_ "github.com/go-chi/chi/middleware"
	"github.com/keycubasic/shortener/internal/dto"
	"github.com/keycubasic/shortener/internal/loader"
	"github.com/keycubasic/shortener/internal/middleware"
	log "github.com/sirupsen/logrus"
)

type application interface {
	GetLink(key string) (string, bool)
	Count() int
	CreateShortLink(link string) (string, error)
}

type configuration interface {
	GetAddress() string
	SetAddress(address string)
	GetBaseURL() string
	SetBaseURL(baseURL string)
	GetFilePath() string
}

type URLHandler struct {
	app      application
	cfg      configuration
	producer *loader.Producer
}

func newURLHandler(app application, cfg configuration) *URLHandler {
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

func (h *URLHandler) closeProducer() error {
	if h.producer != nil {
		if err := h.producer.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (h *URLHandler) saveToFile(key, link string) error {
	if h.producer != nil {
		id := h.app.Count() + 1

		if err := h.producer.Save(id, key, link); err != nil {
			return err
		}
	}

	return nil
}

func (h *URLHandler) getOriginalURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	link, ok := h.app.GetLink(id)

	if !ok {
		http.Error(w, "target link not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", link)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *URLHandler) shortenURL(w http.ResponseWriter, r *http.Request) {
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

	if err := h.saveToFile(key, link); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	resBody := h.cfg.GetBaseURL() + "/" + key

	if _, err := w.Write([]byte(resBody)); err != nil {
		http.Error(w, "unexpected error occurred. try again", http.StatusBadRequest)
		return
	}
}

func (h *URLHandler) shortenURLJson(w http.ResponseWriter, r *http.Request) {
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

	if err := h.saveToFile(key, req.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	r.Get("/{id}", h.getOriginalURL)
	r.Post("/", h.shortenURL)
	r.Post("/api/shorten", h.shortenURLJson)
	return r
}

func RunServer(app application, cfg configuration) {
	h := newURLHandler(app, cfg)

	if err := h.initProducer(); err != nil {
		log.Fatal(err)
	}

	defer h.closeProducer()

	r := h.getRouter()
	log.Infof("Server listening address %s with base URL %s and disk store support at %s", cfg.GetAddress(), cfg.GetBaseURL(), cfg.GetFilePath())

	if err := http.ListenAndServe(cfg.GetAddress(), r); err != nil {
		log.Fatal(err)
	}

	log.Println("Server stops")
}

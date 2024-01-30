package handler

import (
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/keycubasic/shortener/internal/dto"
	"github.com/keycubasic/shortener/internal/middleware"
	json "github.com/mailru/easyjson"
	log "github.com/sirupsen/logrus"
)

const (
	contentTypeJSON = "application/json"
	contentTypeText = "text/plain; charset=utf-8"
)

type application interface {
	GetLink(key string) (string, bool)
	CreateShortLink(link string) (string, error)
}

type configuration interface {
	GetAddress() string
	GetBaseURL() string
	GetFilePath() string
}

type URLHandler struct {
	app application
	cfg configuration
}

func newURLHandler(app application, cfg configuration) *URLHandler {
	return &URLHandler{app: app, cfg: cfg}
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

	w.Header().Add("Content-Type", contentTypeText)
	w.WriteHeader(http.StatusCreated)
	resBody := h.cfg.GetBaseURL() + "/" + key

	if _, err := w.Write([]byte(resBody)); err != nil {
		http.Error(w, "unexpected error occurred. try again", http.StatusBadRequest)
		return
	}
}

func (h *URLHandler) shortenURLJSON(w http.ResponseWriter, r *http.Request) {
	req := dto.NewRequest()
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if err := json.Unmarshal(body, req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	key, err := h.app.CreateShortLink(req.URL)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := dto.NewResponse(h.cfg.GetBaseURL() + "/" + key)
	resBody, err := json.Marshal(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", contentTypeJSON)
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
	r.Post("/api/shorten", h.shortenURLJSON)
	return r
}

func RunServer(app application, cfg configuration) {
	h := newURLHandler(app, cfg)
	r := h.getRouter()
	log.Infof("Server listening address %s with base URL %s and disk store support at %s",
		cfg.GetAddress(), cfg.GetBaseURL(), cfg.GetFilePath())

	if err := http.ListenAndServe(cfg.GetAddress(), r); err != nil {
		log.Fatal(err)
	}

	log.Println("Server stops")
}

package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/keycubasic/shortener/internal/dto"
	"github.com/keycubasic/shortener/internal/middleware"
	"github.com/keycubasic/shortener/internal/storage"
	json "github.com/mailru/easyjson"
	log "github.com/sirupsen/logrus"
)

const (
	contentTypeJSON = "application/json"
	contentTypeText = "text/plain; charset=utf-8"
)

type application interface {
	Ping(ctx context.Context) error
	GetLink(ctx context.Context, key string) (string, error)
	CreateShortLink(ctx context.Context, link string) (string, error)
	GetShortLink(ctx context.Context, link string) (string, error)
	CreateShortLinksBatch(ctx context.Context, linksDto dto.LinksDto) (dto.LinksDto, error)
}

type Cfg interface {
	GetAddress() string
	GetBaseURL() string
	GetFilePath() string
	GetDatabaseDSN() string
}

type URLHandler struct {
	app application
	cfg Cfg
}

func newURLHandler(ctx context.Context, app application, cfg Cfg) *URLHandler {
	return &URLHandler{app: app, cfg: cfg}
}

func (h *URLHandler) getPingDB(w http.ResponseWriter, r *http.Request) {
	if err := h.app.Ping(context.Background()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *URLHandler) getOriginalURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	link, err := h.app.GetLink(context.Background(), id)

	if err != nil {
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

	key, err := h.app.CreateShortLink(context.Background(), link)

	if err != nil && !errors.Is(err, storage.ErrUniqueViolation) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	statusCode := http.StatusCreated

	if errors.Is(err, storage.ErrUniqueViolation) {
		key, err = h.app.GetShortLink(context.Background(), link)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		statusCode = http.StatusConflict
		// w.Header().Add("Content-Encoding", "gzip")
	}

	resBody := h.cfg.GetBaseURL() + "/" + key
	w.Header().Add("Content-Type", contentTypeText)
	w.WriteHeader(statusCode)

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

	key, err := h.app.CreateShortLink(context.Background(), req.URL)

	if err != nil && !errors.Is(err, storage.ErrUniqueViolation) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	statusCode := http.StatusCreated

	if errors.Is(err, storage.ErrUniqueViolation) {
		key, err = h.app.GetShortLink(context.Background(), req.URL)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		statusCode = http.StatusConflict
		// w.Header().Add("Content-Encoding", "gzip")
	}

	res := dto.NewResponse(h.cfg.GetBaseURL() + "/" + key)
	resBody, err := json.Marshal(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", contentTypeJSON)
	w.WriteHeader(statusCode)

	if _, err := w.Write([]byte(resBody)); err != nil {
		http.Error(w, "unexpected error occurred. try again", http.StatusBadRequest)
		return
	}
}

func (h *URLHandler) shortenURLJSONBatch(w http.ResponseWriter, r *http.Request) {
	reqLinksDto := dto.NewLinksDto()
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if err := json.Unmarshal(body, &reqLinksDto); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resLinksDto, err := h.app.CreateShortLinksBatch(context.Background(), reqLinksDto)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resBody, err := json.Marshal(resLinksDto)

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
	r.Get("/ping", h.getPingDB)
	r.Get("/{id}", h.getOriginalURL)
	r.Post("/", h.shortenURL)
	r.Post("/api/shorten", h.shortenURLJSON)
	r.Post("/api/shorten/batch", h.shortenURLJSONBatch)
	return r
}

func RunServer(ctx context.Context, app application, cfg Cfg) {
	h := newURLHandler(ctx, app, cfg)
	r := h.getRouter()
	log.Infof("Server listening address %s with base URL %s",
		cfg.GetAddress(), cfg.GetBaseURL())

	if err := http.ListenAndServe(cfg.GetAddress(), r); err != nil {
		log.Fatal(err)
	}

	log.Println("Server stops")
}

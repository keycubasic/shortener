package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi"
	jsoniter "github.com/json-iterator/go"
	"github.com/keycubasic/shortener/internal/auth"
	"github.com/keycubasic/shortener/internal/dto"
	"github.com/keycubasic/shortener/internal/middleware"
	"github.com/keycubasic/shortener/internal/storage"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var (
	tokenAuth *auth.JWTAuth
	json      = jsoniter.ConfigCompatibleWithStandardLibrary
)

const (
	contentTypeJSON = "application/json"
	contentTypeText = "text/plain; charset=utf-8"

	unexpectedError = "unexpected error occurred. try again"
)

type configuration interface {
	GetAddress() string
	GetBaseURL() string
	GetFilePath() string
	GetSecretKey() string
	GetBufferSize() int
	GetTickerSeconds() int
	GetClosingDuration() int
}

type application interface {
	Ping(ctx context.Context) error
	GetLink(ctx context.Context, key string) (string, error)
	CreateShortLink(ctx context.Context, userID int, link string) (string, error)
	GetShortLink(ctx context.Context, link string) (string, error)
	CreateShortLinksBatch(ctx context.Context, userID int, linksDto dto.LinksDto) (dto.LinksDto, error)
	AddUser(ctx context.Context) (int, error)
	GetUserURLs(ctx context.Context, userID int) (dto.LinksDto, error)
	AppendChan(ctx context.Context, userID int, links dto.Slice)
}

type URLHandler struct {
	cfg configuration
	app application
}

func newURLHandler(cfg configuration, app application) *URLHandler {
	tokenAuth = auth.New("HS256", []byte(cfg.GetSecretKey()), nil)
	return &URLHandler{cfg: cfg, app: app}
}

func (h *URLHandler) pingDB(w http.ResponseWriter, r *http.Request) {
	if err := h.app.Ping(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *URLHandler) getOriginalURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	link, err := h.app.GetLink(r.Context(), id)
	status := http.StatusTemporaryRedirect

	if err != nil && !errors.Is(err, storage.ErrGone) {
		http.Error(w, "target link not found", http.StatusBadRequest)
		return
	}

	if errors.Is(err, storage.ErrGone) {
		status = http.StatusGone
	} else {
		w.Header().Set("Location", link)
	}

	w.WriteHeader(status)
}

func (h *URLHandler) getUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	linksDto, err := h.app.GetUserURLs(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(linksDto) == 0 {
		http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
		return
	}

	resBody, err := json.Marshal(linksDto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(resBody)); err != nil {
		http.Error(w, unexpectedError, http.StatusBadRequest)
		return
	}
}

func (h *URLHandler) shortenURL(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	key, err := h.app.CreateShortLink(r.Context(), userID, link)

	if err != nil && !errors.Is(err, storage.ErrUniqueViolation) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	statusCode := http.StatusCreated

	if errors.Is(err, storage.ErrUniqueViolation) {
		key, err = h.app.GetShortLink(r.Context(), link)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		statusCode = http.StatusConflict
	}

	resBody := strings.Join([]string{h.cfg.GetBaseURL(), key}, "/")
	w.Header().Add("Content-Type", contentTypeText)
	w.WriteHeader(statusCode)

	if _, err := w.Write([]byte(resBody)); err != nil {
		http.Error(w, unexpectedError, http.StatusBadRequest)
		return
	}
}

func (h *URLHandler) shortenURLJSON(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req := dto.NewRequest()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	key, err := h.app.CreateShortLink(r.Context(), userID, req.URL)

	if err != nil && !errors.Is(err, storage.ErrUniqueViolation) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	statusCode := http.StatusCreated

	if errors.Is(err, storage.ErrUniqueViolation) {
		key, err = h.app.GetShortLink(r.Context(), req.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		statusCode = http.StatusConflict
	}

	res := dto.NewResponse(strings.Join([]string{h.cfg.GetBaseURL(), key}, "/"))
	resBody, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", contentTypeJSON)
	w.WriteHeader(statusCode)

	if _, err := w.Write([]byte(resBody)); err != nil {
		http.Error(w, unexpectedError, http.StatusBadRequest)
		return
	}
}

func (h *URLHandler) shortenURLJSONBatch(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reqLinksDto := make(dto.LinksDto, 0)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &reqLinksDto); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resLinksDto, err := h.app.CreateShortLinksBatch(r.Context(), userID, reqLinksDto)
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
		http.Error(w, unexpectedError, http.StatusBadRequest)
		return
	}
}

func (h *URLHandler) deleteUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slice := new(dto.Slice)
	err = json.Unmarshal(body, slice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.app.AppendChan(r.Context(), userID, *slice)

	w.Header().Add("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusAccepted)
}

func (h *URLHandler) getUserID(w http.ResponseWriter, r *http.Request) (userID int, err error) {
	ctx := r.Context()
	token, claims, err := auth.FromContext(ctx)
	if err != nil {
		return 0, fmt.Errorf("handler get user id from context: %w", err)
	}

	if token == nil {
		userID, err = h.app.AddUser(ctx)
		if err != nil {
			return 0, fmt.Errorf("handler get user id add user: %w", err)
		}

		claims := map[string]any{"user_id": userID}
		_, tokenString, err := tokenAuth.Encode(claims)
		if err != nil {
			return 0, fmt.Errorf("handler get user id encode claims: %w", err)
		}

		cookie := http.Cookie{
			Name:  "jwt",
			Value: tokenString,
		}

		http.SetCookie(w, &cookie)
	} else {
		userIDfloat64, ok := claims["user_id"].(float64)

		if !ok {
			return 0, fmt.Errorf("handler get user id convert claims")
		}

		userID = int(userIDfloat64)
	}

	return
}

func (h *URLHandler) authUser(r *http.Request) (userID int, err error) {
	token, claims, err := auth.FromContext(r.Context())
	if err != nil {
		return 0, fmt.Errorf("handler auth user from context: %w", err)
	}

	if token == nil {
		return 0, fmt.Errorf("handler auth user from context: %w", errors.New("empty token"))
	}

	userIDfloat64, ok := claims["user_id"].(float64)

	if !ok {
		return 0, fmt.Errorf("handler auth user convert claims")
	}

	userID = int(userIDfloat64)

	return
}

func (h *URLHandler) getRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Gzip)
	r.Use(middleware.Logger)
	r.Get("/ping", h.pingDB)
	r.Group(func(r chi.Router) {
		r.Use(auth.Verifier(tokenAuth))
		r.Get("/{id}", h.getOriginalURL)
		r.Get("/api/user/urls", h.getUserURLs)
		r.Post("/", h.shortenURL)
		r.Post("/api/shorten", h.shortenURLJSON)
		r.Post("/api/shorten/batch", h.shortenURLJSONBatch)
		r.Delete("/api/user/urls", h.deleteUserURLs)
	})

	r.Handle("/debug/pprof/*", http.DefaultServeMux)

	return r
}

func RunServer(ctx context.Context, cfg configuration, app application) {
	h := newURLHandler(cfg, app)
	r := h.getRouter()
	s := http.Server{Addr: cfg.GetAddress(), Handler: r}
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return s.ListenAndServe()
	})

	log.Infof("Server listening address %s with base URL %s",
		cfg.GetAddress(), cfg.GetBaseURL())

	ctxWithTimeout, cancel := context.
		WithTimeout(context.Background(), time.Duration(cfg.GetClosingDuration()))
	defer cancel()

	g.Go(func() error {
		<-gCtx.Done()
		log.Info("Server stops")
		return s.Shutdown(ctxWithTimeout)
	})

	if err := g.Wait(); err != nil {
		log.Infof("Exit reason: %s", err)
	}
}

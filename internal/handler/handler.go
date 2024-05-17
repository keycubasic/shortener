// Package handler contains handlers and starts the server.
package handler

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi"
	jsoniter "github.com/json-iterator/go"
	"github.com/keycubasic/shortener/internal/auth"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/dto"
	"github.com/keycubasic/shortener/internal/middleware"
	"github.com/keycubasic/shortener/internal/storage"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"

	_ "github.com/keycubasic/shortener/cmd/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

//	@Title			URL shortener API
//	@Version		1.0
//	@Description	Service for shortening URLs
//	@Tags			Resource

var (
	tokenAuth *auth.JWTAuth
	json      = jsoniter.ConfigCompatibleWithStandardLibrary
	//go:embed hosts/whitelist
	embedWhitelist string
)

const (
	contentTypeJSON = "application/json"
	contentTypeText = "text/plain; charset=utf-8"

	unexpectedError = "unexpected error occurred. try again"
)

// Interface of application
type application interface {
	// Ping checks the availability of the data store.
	Ping(ctx context.Context) error
	// GetLink returns the original link by its shortening.
	GetLink(ctx context.Context, key string) (string, error)
	// CreateShortLink сreates a shortening for a link.
	CreateShortLink(ctx context.Context, userID int, link string) (string, error)
	// GetShortLink returns the shortening of the original link.
	GetShortLink(ctx context.Context, link string) (string, error)
	// CreateShortLinksBatch creates shortenings for links in batch mode.
	CreateShortLinksBatch(ctx context.Context, userID int, linksDto dto.LinksDto) (dto.LinksDto, error)
	// AddUser adds a new user.
	AddUser(ctx context.Context) (int, error)
	// GetUserURLs returns user URLs.
	GetUserURLs(ctx context.Context, userID int) (dto.LinksDto, error)
	// AppendChan sends links marked for deletion to the channel for further processing.
	AppendChan(ctx context.Context, userID int, links dto.Slice)
}

// URLHandler stores interfaces for the configuration and functions of the application.
type URLHandler struct {
	cfg *config.Cfg
	app application
}

func newURLHandler(cfg *config.Cfg, app application) *URLHandler {
	tokenAuth = auth.New("HS256", []byte(cfg.SecretKey), nil)
	return &URLHandler{cfg: cfg, app: app}
}

// pingDB checks the availability of the data store.
//
//	@Summary		ping DB
//	@Description	ping database
//	@Tags			Resource
//	@Success		200
//	@Failure		500
//	@Router			/ping [get]
func (h *URLHandler) pingDB(w http.ResponseWriter, r *http.Request) {
	if err := h.app.Ping(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// getOriginalURL returns the original link by its shortening.
//
//	@Summary		get original URL
//	@Description	get original URL by its shortening
//	@Tags			Resource
//	@Param			id	path	string	true	"Shortening"
//	@Success		307
//	@Failure		400
//	@Failure		410
//	@Router			/{id} [get]
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

// getUserURLs returns user URLs in JSON format.
//
//	@Summary		get user URLs
//	@Description	get user URLs in JSON format
//	@Tags			Resource
//	@Success		200	{object}	dto.LinksDto
//	@Failure		204
//	@Failure		400
//	@Failure		401
//	@Router			/api/user/urls [get]
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

// shortenURL сreates a shortening for a link and returns it.
//
//	@Summary		shorten URL
//	@Description	shorten URL and returns it
//	@Tags			Resource
//	@Accept			plain
//	@Produce		plain
//	@Param			URL	body		string	true	"Original URL"
//	@Success		201	{string}	string	"Shorten URL created"
//	@Failure		400
//	@Failure		409
//	@Router			/ [post]
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

	if _, err = url.ParseRequestURI(link); err != nil {
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

	resBody := strings.Join([]string{h.cfg.BaseURL, key}, "/")
	w.Header().Add("Content-Type", contentTypeText)
	w.WriteHeader(statusCode)

	if _, err := w.Write([]byte(resBody)); err != nil {
		http.Error(w, unexpectedError, http.StatusBadRequest)
		return
	}
}

// shortenURLJSON сreates a shortening for a link and returns it in JSON format.
//
//	@Summary		shorten URL in JSON format
//	@Description	shorten URL and returns it in JSON format
//	@Tags			Resource
//	@Accept			json
//	@Produce		json
//	@Param			URL	body		dto.Request		true	"Original URL in JSON format"
//	@Success		201	{object}	dto.Response	"Shorten URL in JSON format"
//	@Failure		400
//	@Failure		409
//	@Router			/api/shorten [post]
func (h *URLHandler) shortenURLJSON(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req := new(dto.Request)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(body, req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err = url.ParseRequestURI(req.URL); err != nil {
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

	res := new(dto.Response)
	res.SortenURL = strings.Join([]string{h.cfg.BaseURL, key}, "/")
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

// shortenURLJSONBatch сreates a shortenings for a links in batch mode and returns it in JSON format.
//
//	@Summary		shorten URLs in JSON format in batch mode
//	@Description	shorten URLs in batch mode and returns it in JSON format
//	@Tags			Resource
//	@Accept			json
//	@Produce		json
//	@Param			URL	body		dto.LinksDto	true	"Array of Original URLs in JSON format"
//	@Success		201	{object}	dto.LinksDto	"Array of Shorten URLs in JSON format"
//	@Failure		400
//	@Router			/api/shorten/batch [post]
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

	if err = json.Unmarshal(body, &reqLinksDto); err != nil {
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

// deleteUserURLs deletes link-shortening pairs.
//
//	@Summary		delete user URLs in batch mode
//	@Description	delete link-shortening pairs in batch mode
//	@Tags			Resource
//	@Accept			json
//	@Param			URL	body	[]string	true	"Array of shortenings"
//	@Success		202
//	@Failure		400
//	@Failure		401
//	@Router			/api/user/urls [delete]
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

// getUserID returns authorized user ID.
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

		claims = map[string]any{"user_id": userID}
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

// authUser authorizes the user.
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

// getRouter returns the configured router.
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
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(h.cfg.ServerAddress+"/swagger/doc.json"),
	))

	return r
}

// RunServer starts the server.
func RunServer(ctx context.Context, cfg *config.Cfg, app application) {
	h := newURLHandler(cfg, app)
	r := h.getRouter()
	addr := cfg.ServerAddress
	var s http.Server
	g, gCtx := errgroup.WithContext(ctx)
	if cfg.EnableHTTPS {
		log.Info("HTTPS enabled")
		manager := &autocert.Manager{
			Cache:      autocert.DirCache("cache-dir"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(strings.Split(embedWhitelist, "\n")...),
		}
		s := http.Server{Addr: addr, Handler: r, TLSConfig: manager.TLSConfig()}
		g.Go(func() error {
			return s.ListenAndServeTLS("", "")
		})
	} else {
		log.Info("HTTPS disabled")
		s := http.Server{Addr: addr, Handler: r}
		g.Go(func() error {
			return s.ListenAndServe()
		})
	}

	log.Infof("Server listening address %s with base URL %s",
		addr, cfg.BaseURL)

	g.Go(func() error {
		<-gCtx.Done()
		log.Info("Server stops")
		ctxWithTimeout, cancel := context.
			WithTimeout(context.Background(), cfg.ClosingTimeout)
		defer cancel()

		return s.Shutdown(ctxWithTimeout)
	})

	if err := g.Wait(); err != nil {
		log.Infof("Exit reason: %s", err)
	}
}

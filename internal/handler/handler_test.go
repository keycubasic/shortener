package handler

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/storage"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	apiShortenURI = "/api/shorten"
)

func connectDB(ctx context.Context, cfg configuration) (app.Database, error) {
	return storage.NewLinks(ctx, cfg)
}

func configureTestHandler(ctx context.Context, cfg configuration) (*URLHandler, error) {
	db, err := connectDB(ctx, cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer func() {
		err = db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	app := app.NewApp(ctx, cfg, db)
	h := newURLHandler(cfg, app)
	return h, nil
}

func BenchmarkJSONHandler(b *testing.B) {
	cfg := new(config.Cfg)
	cfg.ParseCfg()
	cfg.SetBufferSize(0)
	h, err := configureTestHandler(context.Background(), cfg)
	if err != nil {
		b.Fatal(err)
	}

	body := "{\"url\":\"https://ya.ru\"}"
	r := httptest.NewRequest("POST", cfg.GetBaseURL()+apiShortenURI, strings.NewReader(body))
	r.RequestURI = ""
	r.Header.Set("Content-Type", contentTypeJSON)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		h.shortenURLJSON(w, r)
	}
}

func TestRunServer(t *testing.T) {
	var cookie *http.Cookie
	cfg := new(config.Cfg)
	cfg.ParseCfg()
	h, err := configureTestHandler(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	r := h.getRouter()
	ts := httptest.NewServer(r)

	defer ts.Close()
	t.Logf("Test server listening address %s with base URL %s and disk store support at %s",
		cfg.GetAddress(), cfg.GetBaseURL(), cfg.GetFilePath())

	tests := []struct {
		name          string
		method        string
		uri           string
		contentType   string
		body          string
		expectedCodes []int
		expectedBody  string
	}{
		{
			name:          "TestHandler_00_pingOK",
			method:        http.MethodGet,
			uri:           "/ping",
			expectedCodes: []int{http.StatusOK},
		},
		{
			name:          "TestHandler_01_getUserURLsKO",
			method:        http.MethodGet,
			contentType:   contentTypeJSON,
			uri:           "/api/user/urls",
			expectedCodes: []int{http.StatusUnauthorized},
		},
		{
			name:          "TestHandler_02_shortenURLKO",
			method:        http.MethodPost,
			contentType:   contentTypeText,
			expectedCodes: []int{http.StatusBadRequest},
			expectedBody:  "parse \"\": empty url",
		},
		{
			name:          "TestHandler_03_shortenURLKO",
			method:        http.MethodPost,
			contentType:   contentTypeText,
			body:          "1\n2",
			expectedCodes: []int{http.StatusBadRequest},
			expectedBody:  "parse \"1\\n2\": net/url: invalid control character in URL",
		},
		{
			name:          "TestHandler_04_shortenURLKO",
			method:        http.MethodPost,
			contentType:   contentTypeText,
			body:          "1 2",
			expectedCodes: []int{http.StatusBadRequest},
			expectedBody:  "parse \"1 2\": invalid URI for request",
		},
		{
			name:          "TestHandler_05_shortenURLOK",
			method:        http.MethodPost,
			contentType:   contentTypeText,
			body:          "https://google.com",
			expectedCodes: []int{http.StatusCreated, http.StatusConflict},
			expectedBody:  "http://localhost:8080/mZmevP23",
		},
		{
			name:          "TestHandler_06_shortenURLJSONKO",
			method:        http.MethodPost,
			contentType:   contentTypeJSON,
			uri:           apiShortenURI,
			body:          "{\"url\":\"\"}",
			expectedCodes: []int{http.StatusBadRequest},
		},
		{
			name:          "TestHandler_07_shortenURLJSONKO",
			method:        http.MethodPost,
			contentType:   contentTypeJSON,
			uri:           apiShortenURI,
			body:          "{\"url\":\"1 2\"}",
			expectedCodes: []int{http.StatusBadRequest},
		},
		{
			name:          "TestHandler_08_shortenURLJSONOK",
			method:        http.MethodPost,
			contentType:   contentTypeJSON,
			uri:           apiShortenURI,
			body:          "{\"url\":\"https://ya.ru\"}",
			expectedCodes: []int{http.StatusCreated, http.StatusConflict},
			expectedBody:  "{\"result\":\"http://localhost:8080/6YGS4ZUF\"}",
		},
		{
			name:          "TestHandler_09_shortenURLJSONOK",
			method:        http.MethodPost,
			contentType:   contentTypeJSON,
			uri:           apiShortenURI,
			body:          "{\"url\":\"https://ya.ru\"}",
			expectedCodes: []int{http.StatusCreated, http.StatusConflict},
			expectedBody:  "{\"result\":\"http://localhost:8080/6YGS4ZUF\"}",
		},
		{
			name:          "TestHandler_10_getOriginalURLOK",
			method:        http.MethodGet,
			uri:           "/6YGS4ZUF",
			expectedCodes: []int{http.StatusOK, http.StatusBadGateway},
		},
		{
			name:          "TestHandler_11_getOriginalURLOK",
			method:        http.MethodGet,
			uri:           "/TgIlZU3m",
			expectedCodes: []int{http.StatusBadRequest},
			expectedBody:  "target link not found",
		},
		{
			name:          "TestHandler_12_shortenURLJSONBatchKO",
			method:        http.MethodPost,
			contentType:   contentTypeJSON,
			uri:           apiShortenURI + "/batch",
			body:          "[{\"url\":\"https://ya.ru\"}]",
			expectedCodes: []int{http.StatusBadRequest},
		},
		{
			name:        "TestHandler_13_shortenURLJSONBatchOK",
			method:      http.MethodPost,
			contentType: contentTypeText,
			uri:         apiShortenURI + "/batch",
			body: `[
				{
					"correlation_id": "3",
					"original_url": "https://yahoo.com"
				}
			]`,
			expectedCodes: []int{http.StatusCreated},
		},
		{
			name:          "TestHandler_14_getUserURLsOK",
			method:        http.MethodGet,
			contentType:   contentTypeJSON,
			uri:           "/api/user/urls",
			expectedCodes: []int{http.StatusOK, http.StatusNoContent},
		},
		{
			name:          "TestHandler_15_deleteUserURLsOK",
			method:        http.MethodDelete,
			contentType:   contentTypeText,
			uri:           "/api/user/urls",
			body:          `["6YGS4ZUF","rntA0dVu","Yi66Tx-0","s7L9TR66"]`,
			expectedCodes: []int{http.StatusAccepted},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response *http.Response
			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}

			r := httptest.NewRequest(tt.method, ts.URL+tt.uri, body)
			r.RequestURI = ""
			r.Header.Set("Content-Type", tt.contentType)
			r.Header.Set("Accept-Encoding", "gzip")

			if cookie != nil {
				r.AddCookie(cookie)
			}

			res, err := http.DefaultClient.Do(r)
			require.NoError(t, err)

			defer res.Body.Close()
			response = res

			assert.Contains(t, tt.expectedCodes, response.StatusCode, "The response code does not match what is expected")

			if cookie == nil {
				cookies := response.Cookies()
				for _, c := range cookies {
					if c.Name == "jwt" {
						cookie = c
					}
				}
			}

			if tt.expectedBody != "" {
				assert.Equal(t, "gzip", response.Header.Get("Content-Encoding"))
			}

			if tt.expectedBody != "" {
				defer response.Body.Close()
				assert.NotNil(t, response.Body)

				zr, err := gzip.NewReader(response.Body)
				require.NoError(t, err)

				resBody, err := io.ReadAll(zr)
				require.NoError(t, err)

				body := string(resBody)

				if tt.contentType == contentTypeJSON {
					require.JSONEq(t, tt.expectedBody, body)
				}

				assert.Equal(t, tt.expectedBody, strings.TrimSuffix(body, "\n"), "The response body does not match what is expected")
			}
		})
	}
}

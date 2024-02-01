package handler

import (
	"compress/gzip"
	"io"
	"strings"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	apiShortenURI = "/api/shorten"
)

func configureTestServer(cfg configuration) (*httptest.Server, error) {
	dbLegacy, err := storage.NewLinks(cfg)

	if err != nil {
		return nil, err
	}

	app := app.NewApp(cfg, dbLegacy)
	h := newURLHandler(app, cfg)
	r := h.getRouter()
	ts := httptest.NewServer(r)
	return ts, nil
}

func BenchmarkJSONHandler(b *testing.B) {
	cfg := config.NewCfg()
	cfg.ParseCfg()
	dbLegacy, err := storage.NewLinks(cfg)

	if err != nil {
		b.Error(err)
	}

	app := app.NewApp(cfg, dbLegacy)
	h := newURLHandler(app, cfg)
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
	cfg := config.NewCfg()
	cfg.ParseCfg()
	ts, err := configureTestServer(cfg)

	if err != nil {
		t.Log(err)
	}

	defer ts.Close()
	t.Logf("Test server listening address %s with base URL %s and disk store support at %s",
		cfg.GetAddress(), cfg.GetBaseURL(), cfg.GetFilePath())

	tests := []struct {
		name               string
		method             string
		uri                string
		contentType        string
		body               string
		expectedCode       int
		expectedbLegacyody string
	}{
		{
			name:               "TestHandler1",
			method:             http.MethodPost,
			contentType:        contentTypeText,
			expectedCode:       http.StatusBadRequest,
			expectedbLegacyody: "parse \"\": empty url",
		},
		{
			name:               "TestHandler2",
			method:             http.MethodPost,
			contentType:        contentTypeText,
			body:               "1\n2",
			expectedCode:       http.StatusBadRequest,
			expectedbLegacyody: "parse \"1\\n2\": net/url: invalid control character in URL",
		},
		{
			name:               "TestHandler3",
			method:             http.MethodPost,
			contentType:        contentTypeText,
			body:               "1 2",
			expectedCode:       http.StatusBadRequest,
			expectedbLegacyody: "parse \"1 2\": invalid URI for request",
		},
		{
			name:               "TestHandler4",
			method:             http.MethodPost,
			contentType:        contentTypeText,
			body:               "https://google.com",
			expectedCode:       http.StatusCreated,
			expectedbLegacyody: "http://localhost:8080/mZmevP23",
		},
		{
			name:               "TestHandler5",
			method:             http.MethodPost,
			contentType:        contentTypeJSON,
			uri:                apiShortenURI,
			body:               "{\"url\":\"\"}",
			expectedCode:       http.StatusBadRequest,
			expectedbLegacyody: "parse \"\": empty url",
		},
		{
			name:               "TestHandler6",
			method:             http.MethodPost,
			contentType:        contentTypeJSON,
			uri:                apiShortenURI,
			body:               "{\"url\":\"1 2\"}",
			expectedCode:       http.StatusBadRequest,
			expectedbLegacyody: "parse \"1 2\": invalid URI for request",
		},
		{
			name:               "TestHandler7",
			method:             http.MethodPost,
			contentType:        contentTypeJSON,
			uri:                apiShortenURI,
			body:               "{\"url\":\"https://ya.ru\"}",
			expectedCode:       http.StatusCreated,
			expectedbLegacyody: "{\"result\":\"http://localhost:8080/6YGS4ZUF\"}",
		},
		{
			name:         "TestHandler8",
			method:       http.MethodGet,
			uri:          "/6YGS4ZUF",
			expectedCode: http.StatusOK,
		},
		{
			name:               "TestHandler9",
			method:             http.MethodGet,
			uri:                "/TgIlZU3m",
			expectedCode:       http.StatusBadRequest,
			expectedbLegacyody: "target link not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response *http.Response

			switch tt.method {
			case http.MethodGet:
				// res, err := http.Get(ts.URL + tt.uri)
				r := httptest.NewRequest("GET", ts.URL+tt.uri, nil)
				r.RequestURI = ""
				r.Header.Set("Accept-Encoding", "gzip")

				res, err := http.DefaultClient.Do(r)
				require.NoError(t, err)

				defer res.Body.Close()
				response = res

			case http.MethodPost:
				// res, err := http.Post(ts.URL + tt.uri, tt.contentType, strings.NewReader(tt.body))
				r := httptest.NewRequest("POST", ts.URL+tt.uri, strings.NewReader(tt.body))
				r.RequestURI = ""
				r.Header.Set("Content-Type", tt.contentType)
				r.Header.Set("Accept-Encoding", "gzip")

				res, err := http.DefaultClient.Do(r)
				require.NoError(t, err)

				defer res.Body.Close()
				response = res
			}

			assert.Equal(t, tt.expectedCode, response.StatusCode, "The response code does not match what is expected")

			if tt.expectedbLegacyody != "" {
				if response.StatusCode < http.StatusMultipleChoices {
					assert.Equal(t, "gzip", response.Header.Get("Content-Encoding"))
				}

				defer response.Body.Close()
				assert.NotNil(t, response.Body)

				// resBody, err := io.ReadAll(response.Body)
				zr, err := gzip.NewReader(response.Body)
				require.NoError(t, err)

				resBody, err := io.ReadAll(zr)
				require.NoError(t, err)

				body := string(resBody)

				if tt.contentType == contentTypeJSON && response.StatusCode < http.StatusMultipleChoices {
					require.JSONEq(t, tt.expectedbLegacyody, body)
				}

				assert.Equal(t, tt.expectedbLegacyody, strings.TrimSuffix(body, "\n"), "The response body does not match what is expected")
			}
		})
	}
}

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

func TestServer(t *testing.T) {
	address := ":8080"
	baseURL := "http://localhost:8080"
	cfg := config.NewCfg()
	cfg.SetAddress(address)
	cfg.SetBaseURL(baseURL)
	db := storage.NewLinks()
	app := app.NewApp(db)
	h := newURLHandler(app, cfg)
	r := h.getRouter()
	ts := httptest.NewServer(r)
	t.Logf("Test server listening address %s with base URL %s", cfg.GetAddress(), cfg.GetBaseURL())
	defer ts.Close()

	tests := []struct {
		name         string
		method       string
		uri          string
		contentType  string
		body         string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "TestHandler1",
			method:       http.MethodPost,
			contentType:  "text/plain",
			expectedCode: http.StatusBadRequest,
			expectedBody: "parse \"\": empty url",
		},
		{
			name:         "TestHandler2",
			method:       http.MethodPost,
			contentType:  "text/plain",
			body:         "1\n2",
			expectedCode: http.StatusBadRequest,
			expectedBody: "parse \"1\\n2\": net/url: invalid control character in URL",
		},
		{
			name:         "TestHandler3",
			method:       http.MethodPost,
			contentType:  "text/plain",
			body:         "1 2",
			expectedCode: http.StatusBadRequest,
			expectedBody: "parse \"1 2\": invalid URI for request",
		},
		{
			name:         "TestHandler4",
			method:       http.MethodPost,
			contentType:  "text/plain",
			body:         "https://google.com",
			expectedCode: http.StatusCreated,
			expectedBody: "http://localhost:8080/mZmevP23",
		},
		{
			name:         "TestHandler5",
			method:       http.MethodPost,
			contentType:  "application/json",
			uri:          "/api/shorten",
			body:         "https://google.com",
			expectedCode: http.StatusBadRequest,
			expectedBody: "invalid character 'h' looking for beginning of value",
		},
		{
			name:         "TestHandler6",
			method:       http.MethodPost,
			contentType:  "application/json",
			uri:          "/api/shorten",
			body:         "{url:}",
			expectedCode: http.StatusBadRequest,
			expectedBody: "invalid character 'u' looking for beginning of object key string",
		},
		{
			name:         "TestHandler7",
			method:       http.MethodPost,
			contentType:  "application/json",
			uri:          "/api/shorten",
			body:         "{\"url\":\"\"}",
			expectedCode: http.StatusBadRequest,
			expectedBody: "parse \"\": empty url",
		},
		{
			name:         "TestHandler8",
			method:       http.MethodPost,
			contentType:  "application/json",
			uri:          "/api/shorten",
			body:         "{\"url\":\"1\n2\"}",
			expectedCode: http.StatusBadRequest,
			expectedBody: "invalid character '\\n' in string literal",
		},
		{
			name:         "TestHandler9",
			method:       http.MethodPost,
			contentType:  "application/json",
			uri:          "/api/shorten",
			body:         "{\"url\":\"1 2\"}",
			expectedCode: http.StatusBadRequest,
			expectedBody: "url: 1 2 does not validate as url",
		},
		{
			name:         "TestHandler10",
			method:       http.MethodPost,
			contentType:  "application/json",
			uri:          "/api/shorten",
			body:         "{\"url\":\"https://ya.ru\"}",
			expectedCode: http.StatusCreated,
			expectedBody: "{\"result\":\"http://localhost:8080/6YGS4ZUF\"}",
		},
		{
			name:         "TestHandler11",
			method:       http.MethodGet,
			uri:          "/6YGS4ZUF",
			expectedCode: http.StatusOK,
		},
		{
			name:         "TestHandler12",
			method:       http.MethodGet,
			uri:          "/TgIlZU3m",
			expectedCode: http.StatusBadRequest,
			expectedBody: "target link not found",
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

			if tt.expectedBody != "" {
				if response.StatusCode < 300 {
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

				if tt.contentType == "application/json" && response.StatusCode < 300 {
					require.JSONEq(t, tt.expectedBody, body)
				}

				assert.Equal(t, tt.expectedBody, strings.TrimSuffix(body, "\n"), "The response body does not match what is expected")
			}
		})
	}
}

package handler

import (
	"io"
	"strings"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	address := ":8080"
	baseURL := "http://localhost:8080"
	cfg := config.NewCfg()
	cfg.SetAddress(address)
	cfg.SetBaseURL(baseURL)
	db := app.ConnectStorage()
	h := NewURLHandler(cfg, db)
	r := getRouter()
	r.Get("/{id}", h.GetOriginalURL)
	r.Post("/", h.ShortenURL)
	ts := httptest.NewServer(r)
	t.Logf("Test server listening address %s with base URL %s", cfg.GetAddress(), cfg.GetBaseURL())
	defer ts.Close()

	tests := []struct {
		method       string
		uri          string
		body         string
		expectedCode int
		expectedBody string
	}{
		{
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			expectedBody: "body should't be empty",
		},
		{
			method:       http.MethodPost,
			body:         "1\n2",
			expectedCode: http.StatusBadRequest,
			expectedBody: "body contains too many lines",
		},
		{
			method:       http.MethodPost,
			body:         "1 2",
			expectedCode: http.StatusBadRequest,
			expectedBody: "parse \"1 2\": invalid URI for request",
		},
		{
			method:       http.MethodPost,
			body:         "https://google.com",
			expectedCode: http.StatusCreated,
			expectedBody: "http://localhost:8080/mZmevP23",
		},
		{
			method:       http.MethodGet,
			uri:          "/TgIlZU3m",
			expectedCode: http.StatusBadRequest,
			expectedBody: "target link not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			var response *http.Response
			switch tt.method {
			case http.MethodGet:
				res, err := http.Get(ts.URL + tt.uri)
				require.NoError(t, err)
				defer res.Body.Close()
				response = res
			case http.MethodPost:
				res, err := http.Post(ts.URL, "text/plain", strings.NewReader(tt.body))
				require.NoError(t, err)
				defer res.Body.Close()
				response = res
			}
			assert.Equal(t, tt.expectedCode, response.StatusCode, "The response code does not match what is expected")
			if tt.expectedBody != "" {
				defer response.Body.Close()
				assert.NotNil(t, response.Body)
				resBody, _ := io.ReadAll(response.Body)
				body := string(resBody)
				assert.Equal(t, tt.expectedBody, strings.TrimSuffix(body, "\n"), "The response body does not match what is expected")
			}
		})
	}
}

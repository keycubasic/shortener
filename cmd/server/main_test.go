package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	links "github.com/keycubasic/shortener/internal/app"
	"github.com/stretchr/testify/assert"
)

func Test_getHandler(t *testing.T) {
	tests := []struct {
		method       string
		uri          string
		expectedCode int
		expectedBody string
	}{
		{
			method:       http.MethodGet,
			uri:          "/TgIlZU3m",
			expectedCode: http.StatusBadRequest,
			expectedBody: "target link not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.uri, nil)
			w := httptest.NewRecorder()

			getHandler(w, r)
			assert.Equal(t, tt.expectedCode, w.Code, "The response code does not match what is expected")
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, strings.TrimSuffix(w.Body.String(), "\n"), "The response body does not match what is expected")
			}
		})
	}
}

func Test_postHandler(t *testing.T) {
	tests := []struct {
		method       string
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
			body:         "https://practicum.yandex.ru",
			expectedCode: http.StatusBadRequest, // http.StatusInternalServerError
			expectedBody: "shortlink a9tbDian already used for URL https://yandex.ru",
		},
		{
			method:       http.MethodPost,
			body:         "https://google.com",
			expectedCode: http.StatusCreated,
			expectedBody: "http://localhost:8080/mZmevP23",
		},
	}

	links.SetLinkByKey("a9tbDian", "https://yandex.ru") // Wrong shortlink

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			postHandler(w, r)
			assert.Equal(t, tt.expectedCode, w.Code, "The response code does not match what is expected")
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, strings.TrimSuffix(w.Body.String(), "\n"), "The response body does not match what is expected")
			}
		})
	}
}

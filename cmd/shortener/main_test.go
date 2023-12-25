package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_mainHandler(t *testing.T) {
	tests := []struct {
		method       string
		uri	string	
		body string	
        expectedCode int
        expectedBody string
    }{
        {
			method: http.MethodPut, 
			expectedCode: http.StatusBadRequest, // http.StatusMethodNotAllowed, 
			expectedBody: "Method PUT not supported\n",
		},
		{
			method: http.MethodPost,
			expectedCode: http.StatusBadRequest,
			expectedBody: "Boby should't be empty\n",
		},
		{
			method: http.MethodPost,
			body: "1\n2",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Boby contains too many lines\n",
		},
		{
			method: http.MethodPost,
			body: "https://practicum.yandex.ru",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Shortlink a9tbDian already used for URL https://yandex.ru\n",
		},
		{
			method: http.MethodPost,
			body: "https://google.com",
			expectedCode: http.StatusCreated,
			expectedBody: "example.com/mZmevP23",
		},
		{
			method: http.MethodGet,
			uri: "/mZmevP23/qwerty",
			expectedCode: http.StatusBadRequest,
			expectedBody: "too many parameters in source path /mZmevP23/qwerty\n",
		},
		{
			method: http.MethodGet,
			uri: "/TgIlZU3m",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Source link not found\n",
		},
		{
			method: http.MethodGet,
			uri: "/mZmevP23",
			expectedCode: http.StatusTemporaryRedirect,
			expectedBody: "",
		},
	}

	shortLinks["a9tbDian"] = "https://yandex.ru" // Wrong shortlink
	
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			if tt.uri == "" {
				tt.uri = "/" 
			}
			r := httptest.NewRequest(tt.method, tt.uri, strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			mainHandler(w, r)
			assert.Equal(t, tt.expectedCode, w.Code, "The response code does not match what is expected")
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String(), "The response body does not match what is expected")
			}
		})
	}
}

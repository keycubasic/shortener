package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	links "github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/octago/sflags/gen/gflag"
)

var cfg = &config.Params{
	A: ":8080",
	B: "http://localhost:8080/",
}

func main() {
	if err := gflag.ParseToDef(cfg); err != nil {
		log.Fatalln(err)
	}

	flag.Parse()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/{id}", getHandler)
	r.Post("/", postHandler)
	log.Println("Starting server..")
	if err := http.ListenAndServe(cfg.A, r); err != nil {
		log.Fatal(err)
	}
	log.Println("Server starts on port", cfg.A)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	link, ok := links.GetLinkByKey(id)
	if !ok {
		http.Error(w, "target link not found", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", link)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if body, err := io.ReadAll(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		link := string(body)
		if link == "" {
			http.Error(w, "body should't be empty", http.StatusBadRequest)
			return
		}
		if strings.Contains(link, "\n") {
			http.Error(w, "body contains too many lines", http.StatusBadRequest)
			return
		}
		if _, err := url.ParseRequestURI(link); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		key, err := links.CreateShortLink(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest) // http.StatusInternalServerError
			return
		}
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		resBody := cfg.B + key
		w.Write([]byte(resBody))
	}
}

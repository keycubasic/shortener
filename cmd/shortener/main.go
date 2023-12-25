package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	utils "github.com/keycubasic/shortener/internal/app"
)

var shortLinks = make(map[string]string)

func mainHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if key, err := utils.GetKeyFromPath(r.URL.Path); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			link, ok := shortLinks[key]
			if !ok {
				http.Error(w, "Source link not found", http.StatusBadRequest)
				return
			}
			w.Header().Set("Location", link)
			w.WriteHeader(http.StatusTemporaryRedirect)			
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Can't parse request body", http.StatusBadRequest)
			return
		}
		if body, err := io.ReadAll(r.Body); err != nil {
			http.Error(w, "Can't read request body", http.StatusBadRequest)
			return
		} else {
			link := string(body) 
			if link == "" {
				http.Error(w, "Boby should't be empty", http.StatusBadRequest)
				return
			}
			if strings.Contains(link, "\n") {
				http.Error(w, "Boby contains too many lines", http.StatusBadRequest)
				return
			}
			/*if _, err := url.ParseRequestURI(link); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}*/
			key, _ := utils.CreateShortLink(body)
			if val, ok := shortLinks[key]; ok && val != link {
				http.Error(w, fmt.Sprintf("Shortlink %s already used for URL %s", key, val), http.StatusBadRequest) // http.StatusInternalServerError
				return
			}
			shortLinks[key] = link
			w.Header().Add("Content-Type", "text/plain")
			w.WriteHeader(http.StatusCreated)
			resBody := r.Host + r.URL.Port() + "/" + key + "\n"
			w.Write([]byte(resBody))
		}
	default:
		http.Error(w, fmt.Sprintf("Method %s not supported", r.Method), http.StatusBadRequest) //http.StatusMethodNotAllowed
		return
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", mainHandler)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

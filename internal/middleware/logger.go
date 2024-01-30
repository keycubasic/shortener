package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		t := time.Now()
		defer func() {
			log.WithFields(log.Fields{
				"remote": r.RemoteAddr,
				"status": ww.Status(),
				"size":   ww.BytesWritten(),
				"time":   time.Since(t),
			}).Infof("%s %s%s %s", r.Method, r.Host, r.RequestURI, r.Proto)
		}()

		next.ServeHTTP(ww, r)
	})
}

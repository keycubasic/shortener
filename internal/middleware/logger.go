package middleware

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type logRecord struct {
	http.ResponseWriter
	status int
	size   int
}

func newLogRecord(w http.ResponseWriter) *logRecord {
	return &logRecord{
		ResponseWriter: w,
		status:         0,
		size:           0,
	}
}

func (r *logRecord) Write(p []byte) (int, error) {
	size, err := r.ResponseWriter.Write(p)
	r.size += size
	return size, err
}

func (r *logRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		rec := newLogRecord(w)

		defer func() {
			log.WithFields(log.Fields{
				"remote": r.RemoteAddr,
				"status": rec.status,
				"size":   rec.size,
				"time":   time.Since(t),
			}).Infof("%s %s%s %s", r.Method, r.Host, r.RequestURI, r.Proto)
		}()

		next.ServeHTTP(rec, r)
	})
}

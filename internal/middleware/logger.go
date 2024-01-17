package middleware

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type LogRecord struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *LogRecord) Write(p []byte) (int, error) {
	size, err := r.ResponseWriter.Write(p)
	r.size += size
	return size, err
}

func (r *LogRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		record := &LogRecord{
			ResponseWriter: w,
			status:         0,
			size:           0,
		}
		defer func() {
			log.WithFields(log.Fields{
				"Remote": r.RemoteAddr,
				"Status": record.status,
				"Size":   record.size,
				"Time":   time.Since(t1),
			}).Infof("%s %s%s %s", r.Method, r.Host, r.RequestURI, r.Proto)
		}()
		next.ServeHTTP(record, r)
	})
}

// Package middleware contains middleware for compress and decompress data.
package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// compressWriter
type compressWriter struct {
	http.ResponseWriter
	zw *gzip.Writer
}

// newCompressWriter
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		zw:             gzip.NewWriter(w),
	}
}

// Header
func (c *compressWriter) Header() http.Header {
	return c.ResponseWriter.Header()
}

// Write
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader
func (c *compressWriter) WriteHeader(statusCode int) {
	c.Header().Set("Content-Encoding", "gzip")
	c.ResponseWriter.WriteHeader(statusCode)
}

// Close
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader
type compressReader struct {
	io.ReadCloser
	zr *gzip.Reader
}

// newCompressReader
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("compressor create reader: %w", err)
	}

	return &compressReader{
		ReadCloser: r,
		zr:         zr,
	}, nil
}

// Read
func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close
func (c *compressReader) Close() error {
	if err := c.ReadCloser.Close(); err != nil {
		return fmt.Errorf("compressor close reader: %w", err)
	}

	return c.zr.Close()
}

// Gzip provides intermediate compression and decompression.
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw

			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")

		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			r.Body = cr

			defer cr.Close()
		}

		next.ServeHTTP(ow, r)
	})
}

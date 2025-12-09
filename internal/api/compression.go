package api

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// gzipResponseWriter wraps http.ResponseWriter to provide gzip compression
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

// Write compresses data before writing
func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// withGzip wraps a handler with gzip compression
func (s *Server) withGzip(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// Client doesn't support gzip, serve normally
			next(w, r)
			return
		}

		// Don't compress file downloads (already compressed or binary)
		if strings.HasPrefix(r.URL.Path, "/download/") {
			next(w, r)
			return
		}

		// Don't compress already compressed static files
		if strings.HasSuffix(r.URL.Path, ".jpg") ||
			strings.HasSuffix(r.URL.Path, ".jpeg") ||
			strings.HasSuffix(r.URL.Path, ".png") ||
			strings.HasSuffix(r.URL.Path, ".gif") ||
			strings.HasSuffix(r.URL.Path, ".zip") ||
			strings.HasSuffix(r.URL.Path, ".gz") {
			next(w, r)
			return
		}

		// Set headers for gzip compression
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		// Create gzip writer
		gz := gzip.NewWriter(w)
		defer gz.Close()

		// Wrap response writer
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}

		// Call next handler with gzip writer
		next(gzw, r)
	}
}

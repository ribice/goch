// Package httplog is a fork of https://github.com/unrolled/logger
package httplog

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// Service represents http request/response log middlware service
type Service struct {
	log         *log.Logger
	prefix      string
	ignoredURIs []string
}

// New returns a new Logger instance.
func New(prefix string, ingoredURIs ...string) *Service {
	return &Service{
		log:         log.New(os.Stdout, prefix, log.LstdFlags),
		prefix:      prefix,
		ignoredURIs: ingoredURIs,
	}
}

var xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
var xRealIP = http.CanonicalHeaderKey("X-Real-IP")

// MWFunc wraps an HTTP handler and logs the request as necessary.
func (s *Service) MWFunc(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		crw := newCustomResponseWriter(w)
		h.ServeHTTP(crw, r)

		for _, uri := range s.ignoredURIs {
			if uri == r.RequestURI {
				return
			}
		}

		if xff := r.Header.Get(xForwardedFor); xff != "" {
			i := strings.Index(xff, ", ")
			if i == -1 {
				i = len(xff)
			}
			r.RemoteAddr = xff[:i]
		} else if xrip := r.Header.Get(xRealIP); xrip != "" {
			r.RemoteAddr = xrip
		}

		s.log.Printf("(%s) \"%s %s %s\" %d %d %s", r.RemoteAddr, r.Method, r.RequestURI, r.Proto, crw.status, crw.size, time.Since(start))
	})
}

type customResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (c *customResponseWriter) WriteHeader(status int) {
	c.status = status
	c.ResponseWriter.WriteHeader(status)
}

func (c *customResponseWriter) Write(b []byte) (int, error) {
	size, err := c.ResponseWriter.Write(b)
	c.size += size
	return size, err
}

func (c *customResponseWriter) Flush() {
	if f, ok := c.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (c *customResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := c.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, errors.New("ResponseWriter does not implement the Hijacker interface")
}

func newCustomResponseWriter(w http.ResponseWriter) *customResponseWriter {
	// When WriteHeader is not called, it's safe to assume the status will be 200.
	return &customResponseWriter{
		ResponseWriter: w,
	}
}

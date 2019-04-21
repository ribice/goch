// Package recovery is a fork of https://github.com/unrolled/recovery
package recovery

import (
	"log"
	"net/http"
	"os"
	"runtime"
)

// Service represents recovery middleware service
type Service struct {
	*log.Logger
	panicHandler http.Handler
	prefix       string
}

// New returns a new Recovery instance.
func New(prefix string) *Service {
	return &Service{
		Logger:       log.New(os.Stderr, prefix, log.LstdFlags),
		panicHandler: http.HandlerFunc(defaultPanicHandler),
	}
}

// MWFunc wraps an HTTP handler and recovers any panics from upstream.
func (s *Service) MWFunc(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.panicHandler.ServeHTTP(w, req)

				stack := make([]byte, 8096)
				stack = stack[:runtime.Stack(stack, false)]

				s.Printf("Recovering from Panic: %s\n%s", err, stack)
			}
		}()

		h.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}

// SetPanicHandler sets the handler to call when Recovery encounters a panic.
func (s *Service) SetPanicHandler(handler http.Handler) {
	s.panicHandler = handler
}

func defaultPanicHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

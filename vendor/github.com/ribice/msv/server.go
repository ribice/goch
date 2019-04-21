package msv

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ribice/msv/middleware/httplog"
	"github.com/ribice/msv/middleware/recovery"

	"github.com/gorilla/mux"
)

// Server represents http server
type Server struct {
	m *mux.Router
	*http.Server
}

// New instantiates new http server with logging and recover middleware
func New(port int, prefix string) (*Server, *mux.Router) {
	m := mux.NewRouter()
	rmw := recovery.New(prefix)
	lmw := httplog.New(prefix, "/")
	m.Use(rmw.MWFunc, lmw.MWFunc)

	srv := &Server{m: m, Server: &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: m,
	}}
	return srv, srv.m
}

// Start starts the http server
func (s *Server) Start() error {
	go func() {
		log.Printf("starting server on port%v", s.Addr)
		s.ListenAndServe()
	}()
	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (pkill -2)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		return fmt.Errorf("error stopping server: %s", err)
	}

	log.Print("gracefully stopped server")
	return nil
}

// StartTLS starts the https server
func (s *Server) StartTLS(cf, kf string) error {
	go func() {
		log.Printf("starting server on port%v", s.Addr)
		s.ListenAndServeTLS(cf, kf)
	}()
	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (pkill -2)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		return fmt.Errorf("error stopping server: %s", err)
	}

	log.Print("gracefully stopped server")
	return nil
}

package msv

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

// Server represents http server
type Server struct {
	m *mux.Router
	*http.Server
}

// New instantiates new http server
func New(port int) (*Server, *mux.Router) {
	m := mux.NewRouter()
	// m.Use(middleware.Recoverer, middleware.Logger, middleware.RequestID, middleware.RealIP)

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
		if err := s.ListenAndServe(); err != nil {
			log.Printf("error starting server: %s\n", err)
		}
	}()
	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (pkill -2)
	<-stop

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
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
		if err := s.ListenAndServeTLS(cf, kf); err != nil {
			log.Printf("error starting server: %s\n", err)
		}
	}()
	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (pkill -2)
	<-stop

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err := s.Shutdown(ctx); err != nil {
		return fmt.Errorf("error stopping server: %s", err)
	}

	log.Print("gracefully stopped server")
	return nil
}

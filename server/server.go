package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
	"vmuser/ext/httpext/responses"
)

type Config struct {
	Port string
}

type Server struct {
	config *Config
	mux    *http.ServeMux
}

func NewServer(config *Config) *Server {
	return &Server{
		config: config,
		mux:    http.NewServeMux(),
	}
}

func (s *Server) Start(appCtx context.Context) error {
	s.registerRoutes()
	addr := fmt.Sprintf(":%s", s.config.Port)
	log.Printf("Server starting on %s", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: s.mux,
	}

	go func() {
		<-appCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server Shutdown Failed:%+v", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /api/v1/{cmd}", HandlerGeneralCommand())
}

func HandlerGeneralCommand() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd := r.PathValue("cmd")

		response := map[string]interface{}{
			"cmd": cmd,
		}

		responses.JsonOK(w, response)
	}
}

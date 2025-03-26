package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/", s.HelloWorldHandler)

	r.Get("/health", s.healthHandler)

	r.Get("/textract", s.textractHandler)

	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{
		Message: "Hello World",
	}

	s.writeJSON(w, http.StatusAccepted, resp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	health := s.db.Health()
	healthErr, _ := health["error"]
	healthMessage, _ := health["message"]

	resp := jsonResponse{
		Error:   healthErr != "",
		Message: healthMessage,
		Data:    health,
	}

	s.writeJSON(w, http.StatusAccepted, resp)
}

func (s *Server) textractHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	blocks, err := s.extractText(ctx)
	if err != nil {
		s.errorJSON(w, err, http.StatusBadRequest)
	}

	resp := jsonResponse{
		Message: "Extracted text successfully!",
		Data:    blocks,
	}

	s.writeJSON(w, http.StatusAccepted, resp)
}

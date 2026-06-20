package api

import (
	"database/sql"
	"log"
	"net/http"
)

type Server struct {
	DB *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{DB: db}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/transactions", s.handleTransactions)
	mux.HandleFunc("/categories", s.handleCategories)
	mux.HandleFunc("/methods", s.handleMethods)

	handler := enableCORS(mux)
	handler = loggingMiddleware(handler)

	return handler
}

func (s *Server) Run(addr string) error {
	log.Println("🚀 API server running on", addr)
	return http.ListenAndServe(addr, s.Routes())
}

func enableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func loggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log.Printf("➡️ %s %s", r.Method, r.URL.Path)

		h.ServeHTTP(w, r)
	})
}

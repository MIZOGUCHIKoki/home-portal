package api

import (
	"database/sql"
	"log"
	"net/http"

	"kakeibo/internal/storage"
)

type Server struct {
	DB *sql.DB
	S3 *storage.S3Client
}

func NewServer(db *sql.DB, s3Client *storage.S3Client) *Server {
	return &Server{DB: db, S3: s3Client}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/transactions", s.handleTransactions)
	mux.HandleFunc("/advances", s.handleAdvances)

	mux.HandleFunc("/categories", s.handleCategories)
	mux.HandleFunc("/methods", s.handleMethods)
	mux.HandleFunc("/place-rules", s.handlePlaceRules)

	mux.HandleFunc("/transactions/summary/monthly", s.handleMonthlySummary)
	mux.HandleFunc("/transactions/summary/category", s.handleCategorySummary)

	mux.HandleFunc("/csv/upload", s.handleCsvUpload)
	mux.HandleFunc("/csv/check", s.handleCsvCheck)

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

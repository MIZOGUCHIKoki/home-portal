package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

type APIServer struct {
	db  *sql.DB
	mux *http.ServeMux
}

type apiError struct {
	Error string `json:"error"`
}

type userInput struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"isAdmin"`
}

type userResponse struct {
	UserID    int64      `json:"userId"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	IsAdmin   bool       `json:"isAdmin"`
	Login     *time.Time `json:"login,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type categoryInput struct {
	CategoryName string `json:"categoryName"`
}

type categoryResponse struct {
	CategoryID   int64  `json:"categoryId"`
	CategoryName string `json:"categoryName"`
}

type methodInput struct {
	MethodName string `json:"methodName"`
}

type methodResponse struct {
	MethodID   int64  `json:"methodId"`
	MethodName string `json:"methodName"`
}

func StartAPIServer(db *sql.DB) {
	port := getEnv("API_PORT", "8080")
	server := newAPIServer(db)

	log.Printf("🌐 API mode listening on :%s", port)
	if err := http.ListenAndServe(":"+port, server.mux); err != nil {
		log.Fatalf("API server error: %v", err)
	}
}

func newAPIServer(db *sql.DB) *APIServer {
	server := &APIServer{
		db:  db,
		mux: http.NewServeMux(),
	}

	// 共通エンドポイント
	server.mux.HandleFunc("GET /health", server.handleHealth)
	server.mux.HandleFunc("GET /", server.handleRoot)

	// ユーザー管理エンドポイント (認証なし)
	server.mux.HandleFunc("GET /users", server.handleListUsers)
	server.mux.HandleFunc("POST /users", server.handleCreateUser)
	server.mux.HandleFunc("GET /users/{id}", server.handleGetUser)
	server.mux.HandleFunc("PUT /users/{id}", server.handleUpdateUser)
	server.mux.HandleFunc("DELETE /users/{id}", server.handleDeleteUser)

	// カテゴリ管理エンドポイント (認証なし)
	server.mux.HandleFunc("GET /categories", server.handleListCategories)
	server.mux.HandleFunc("POST /categories", server.handleCreateCategory)
	server.mux.HandleFunc("GET /categories/{id}", server.handleGetCategory)
	server.mux.HandleFunc("PUT /categories/{id}", server.handleUpdateCategory)
	server.mux.HandleFunc("DELETE /categories/{id}", server.handleDeleteCategory)

	// 決済方法管理エンドポイント (認証なし)
	server.mux.HandleFunc("GET /methods", server.handleListMethods)
	server.mux.HandleFunc("POST /methods", server.handleCreateMethod)
	server.mux.HandleFunc("GET /methods/{id}", server.handleGetMethod)
	server.mux.HandleFunc("PUT /methods/{id}", server.handleUpdateMethod)
	server.mux.HandleFunc("DELETE /methods/{id}", server.handleDeleteMethod)

	return server
}

func (s *APIServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		notFound(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "BudgetMS API"})
}

func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- ユーザーハンドラー ---

func (s *APIServer) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := ListUsers(s.db)
	if err != nil {
		writeDBError(w, err)
		return
	}

	result := make([]userResponse, 0, len(users))
	for _, user := range users {
		result = append(result, toUserResponse(user))
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *APIServer) handleGetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parsePathID(r.PathValue("id"))
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	user, err := GetUserByID(s.db, userID)
	if err != nil {
		writeDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(*user))
}

func (s *APIServer) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var input userInput
	if err := decodeJSON(r, &input); err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	userID, err := CreateUserWithAdmin(s.db, input.Email, input.Name, input.Password, input.IsAdmin)
	if err != nil {
		writeDBError(w, err)
		return
	}

	user, err := GetUserByID(s.db, userID)
	if err != nil {
		writeDBError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toUserResponse(*user))
}

func (s *APIServer) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parsePathID(r.PathValue("id"))
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	var input userInput
	if err := decodeJSON(r, &input); err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	if err := UpdateUserWithAdmin(s.db, userID, input.Email, input.Name, input.Password, input.IsAdmin); err != nil {
		writeDBError(w, err)
		return
	}

	user, err := GetUserByID(s.db, userID)
	if err != nil {
		writeDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(*user))
}

func (s *APIServer) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parsePathID(r.PathValue("id"))
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	if err := DeleteUser(s.db, userID); err != nil {
		writeDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- カテゴリハンドラー ---

func (s *APIServer) handleListCategories(w http.ResponseWriter, r *http.Request) {
	items, err := ListCategories(s.db)
	if err != nil {
		writeDBError(w, err)
		return
	}

	result := make([]categoryResponse, 0, len(items))
	for _, item := range items {
		result = append(result, categoryResponse{CategoryID: item.CategoryID, CategoryName: item.CategoryName})
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *APIServer) handleGetCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, err := parsePathID(r.PathValue("id"))
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	item, err := GetCategoryByID(s.db, categoryID)
	if err != nil {
		writeDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, categoryResponse{CategoryID: item.CategoryID, CategoryName: item.CategoryName})
}

func (s *APIServer) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	var input categoryInput
	if err := decodeJSON(r, &input); err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	categoryID, err := CreateCategory(s.db, input.CategoryName)
	if err != nil {
		writeDBError(w, err)
		return
	}

	item, err := GetCategoryByID(s.db, categoryID)
	if err != nil {
		writeDBError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, categoryResponse{CategoryID: item.CategoryID, CategoryName: item.CategoryName})
}

func (s *APIServer) handleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, err := parsePathID(r.PathValue("id"))
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	var input categoryInput
	if err := decodeJSON(r, &input); err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	if err := UpdateCategory(s.db, categoryID, input.CategoryName); err != nil {
		writeDBError(w, err)
		return
	}

	item, err := GetCategoryByID(s.db, categoryID)
	if err != nil {
		writeDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, categoryResponse{CategoryID: item.CategoryID, CategoryName: item.CategoryName})
}

func (s *APIServer) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, err := parsePathID(r.PathValue("id"))
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	if err := DeleteCategory(s.db, categoryID); err != nil {
		writeDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- 決済方法ハンドラー ---

func (s *APIServer) handleListMethods(w http.ResponseWriter, r *http.Request) {
	items, err := ListMethods(s.db)
	if err != nil {
		writeDBError(w, err)
		return
	}

	result := make([]methodResponse, 0, len(items))
	for _, item := range items {
		result = append(result, methodResponse{MethodID: item.MethodID, MethodName: item.MethodName})
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *APIServer) handleGetMethod(w http.ResponseWriter, r *http.Request) {
	methodID, err := parsePathID(r.PathValue("id"))
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	item, err := GetMethodByID(s.db, methodID)
	if err != nil {
		writeDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, methodResponse{MethodID: item.MethodID, MethodName: item.MethodName})
}

func (s *APIServer) handleCreateMethod(w http.ResponseWriter, r *http.Request) {
	var input methodInput
	if err := decodeJSON(r, &input); err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	methodID, err := CreateMethod(s.db, input.MethodName)
	if err != nil {
		writeDBError(w, err)
		return
	}

	item, err := GetMethodByID(s.db, methodID)
	if err != nil {
		writeDBError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, methodResponse{MethodID: item.MethodID, MethodName: item.MethodName})
}

func (s *APIServer) handleUpdateMethod(w http.ResponseWriter, r *http.Request) {
	methodID, err := parsePathID(r.PathValue("id"))
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	var input methodInput
	if err := decodeJSON(r, &input); err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	if err := UpdateMethod(s.db, methodID, input.MethodName); err != nil {
		writeDBError(w, err)
		return
	}

	item, err := GetMethodByID(s.db, methodID)
	if err != nil {
		writeDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, methodResponse{MethodID: item.MethodID, MethodName: item.MethodName})
}

func (s *APIServer) handleDeleteMethod(w http.ResponseWriter, r *http.Request) {
	methodID, err := parsePathID(r.PathValue("id"))
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	if err := DeleteMethod(s.db, methodID); err != nil {
		writeDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- ヘルパー関数群 ---

func parsePathID(raw string) (int64, error) {
	if raw == "" {
		return 0, errors.New("id is required")
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("id must be a positive integer")
	}

	return id, nil
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}

func writeBadRequest(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, apiError{Error: message})
}

func writeDBError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, apiError{Error: "not found"})
		return
	}

	log.Printf("db error: %v", err)
	writeJSON(w, http.StatusInternalServerError, apiError{Error: "internal server error"})
}

func notFound(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, apiError{Error: "not found"})
}

func toUserResponse(user User) userResponse {
	response := userResponse{
		UserID:    user.UserID,
		Email:     user.Email,
		Name:      user.Name,
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	if user.Login.Valid {
		response.Login = &user.Login.Time
	}

	return response
}
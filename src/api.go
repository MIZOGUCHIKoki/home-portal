package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"encoding/base64"
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

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token   string       `json:"token"`
	Expires time.Time    `json:"expires"`
	User    userResponse `json:"user"`
}

type authClaims struct {
	UserID  int64  `json:"uid"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"isAdmin"`
	Exp     int64  `json:"exp"`
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

	server.mux.HandleFunc("GET /health", server.handleHealth)
	server.mux.HandleFunc("GET /", server.handleRoot)
	server.mux.HandleFunc("POST /login", server.handleLogin)

	server.mux.HandleFunc("GET /users", server.handleListUsers)
	server.mux.HandleFunc("POST /users", server.withAdmin(server.handleCreateUser))
	server.mux.HandleFunc("GET /users/{id}", server.handleGetUser)
	server.mux.HandleFunc("PUT /users/{id}", server.withAdmin(server.handleUpdateUser))
	server.mux.HandleFunc("DELETE /users/{id}", server.withAdmin(server.handleDeleteUser))

	server.mux.HandleFunc("GET /categories", server.handleListCategories)
	server.mux.HandleFunc("POST /categories", server.withAdmin(server.handleCreateCategory))
	server.mux.HandleFunc("GET /categories/{id}", server.handleGetCategory)
	server.mux.HandleFunc("PUT /categories/{id}", server.withAdmin(server.handleUpdateCategory))
	server.mux.HandleFunc("DELETE /categories/{id}", server.withAdmin(server.handleDeleteCategory))

	server.mux.HandleFunc("GET /methods", server.handleListMethods)
	server.mux.HandleFunc("POST /methods", server.withAdmin(server.handleCreateMethod))
	server.mux.HandleFunc("GET /methods/{id}", server.handleGetMethod)
	server.mux.HandleFunc("PUT /methods/{id}", server.withAdmin(server.handleUpdateMethod))
	server.mux.HandleFunc("DELETE /methods/{id}", server.withAdmin(server.handleDeleteMethod))

	return server
}

func (s *APIServer) withAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := extractBearerToken(r.Header.Get("Authorization"))
		if err != nil {
			writeUnauthorized(w, err.Error())
			return
		}

		claims, err := parseAuthToken(token)
		if err != nil {
			writeUnauthorized(w, "invalid token")
			return
		}
		if !claims.IsAdmin {
			writeForbidden(w, "admin required")
			return
		}

		next(w, r)
	}
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

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	var input loginInput
	if err := decodeJSON(r, &input); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if input.Email == "" || input.Password == "" {
		writeBadRequest(w, "email and password are required")
		return
	}

	user, err := GetUserByEmail(s.db, input.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeUnauthorized(w, "invalid credentials")
			return
		}
		writeDBError(w, err)
		return
	}

	if err := CheckPassword(user.Password, input.Password); err != nil {
		writeUnauthorized(w, "invalid credentials")
		return
	}

	expires := time.Now().Add(getTokenTTL())
	token, err := buildAuthToken(authClaims{
		UserID:  user.UserID,
		Email:   user.Email,
		IsAdmin: user.IsAdmin,
		Exp:     expires.Unix(),
	})
	if err != nil {
		log.Printf("token create error: %v", err)
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "internal server error"})
		return
	}

	if err := markUserLogin(s.db, user.UserID); err != nil {
		log.Printf("user login timestamp update error: %v", err)
	}

	writeJSON(w, http.StatusOK, loginResponse{
		Token:   token,
		Expires: expires,
		User:    toUserResponse(*user),
	})
}

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

func writeUnauthorized(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusUnauthorized, apiError{Error: message})
}

func writeForbidden(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusForbidden, apiError{Error: message})
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

func extractBearerToken(headerValue string) (string, error) {
	if headerValue == "" {
		return "", errors.New("authorization header is required")
	}

	parts := strings.SplitN(headerValue, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", errors.New("authorization header must be Bearer token")
	}

	return parts[1], nil
}

func markUserLogin(db *sql.DB, userID int64) error {
	_, err := db.Exec("UPDATE users SET login = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1", userID)
	return err
}

func getTokenTTL() time.Duration {
	raw := getEnv("AUTH_TOKEN_TTL_HOURS", "24")
	hours, err := strconv.Atoi(raw)
	if err != nil || hours <= 0 {
		hours = 24
	}

	return time.Duration(hours) * time.Hour
}

func getAuthSecret() []byte {
	return []byte(getEnv("AUTH_TOKEN_SECRET", "dev-only-change-me"))
}

func buildAuthToken(claims authClaims) (string, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, getAuthSecret())
	if _, err := mac.Write([]byte(payloadB64)); err != nil {
		return "", err
	}
	signatureB64 := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s.%s", payloadB64, signatureB64), nil
}

func parseAuthToken(token string) (*authClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid token format")
	}

	payloadB64 := parts[0]
	givenSigB64 := parts[1]

	mac := hmac.New(sha256.New, getAuthSecret())
	if _, err := mac.Write([]byte(payloadB64)); err != nil {
		return nil, err
	}
	expectedSig := mac.Sum(nil)

	givenSig, err := base64.RawURLEncoding.DecodeString(givenSigB64)
	if err != nil {
		return nil, err
	}
	if !hmac.Equal(givenSig, expectedSig) {
		return nil, errors.New("invalid token signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, err
	}

	var claims authClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	if claims.UserID <= 0 || claims.Exp <= 0 {
		return nil, errors.New("invalid token claims")
	}
	if time.Now().Unix() > claims.Exp {
		return nil, errors.New("token expired")
	}

	return &claims, nil
}
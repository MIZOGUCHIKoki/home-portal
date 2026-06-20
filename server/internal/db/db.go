package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

func ConnectDB() (*sql.DB, error) {
    host := os.Getenv("DB_HOST")
    if host == "" {
        return nil, fmt.Errorf("DB_HOST is required")
    }

    port := os.Getenv("DB_PORT")
    if port == "" {
        return nil, fmt.Errorf("DB_PORT is required")
    }

    user := os.Getenv("DB_USER")
    if user == "" {
        return nil, fmt.Errorf("DB_USER is required")
    }

    password := os.Getenv("DB_PASSWORD")
    if password == "" {
        return nil, fmt.Errorf("DB_PASSWORD is required")
    }

    dbName := os.Getenv("DB_NAME")
    if dbName == "" {
        return nil, fmt.Errorf("DB_NAME is required")
    }

    sslMode := os.Getenv("DB_SSLMODE")
    if sslMode == "" {
        sslMode = "disable"
    }

    if _, err := strconv.Atoi(port); err != nil {
        return nil, fmt.Errorf("DB_PORT is invalid: %w", err)
    }

    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        host, port, user, password, dbName, sslMode,
    )

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("DB接続設定エラー: %w", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("DB疎通確認エラー: %w", err)
    }

    fmt.Println("📦 DB接続成功")

    return db, nil
}
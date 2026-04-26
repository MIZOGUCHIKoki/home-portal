package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// ConnectDB は環境変数から接続情報を読み取り DB 接続
func ConnectDB() *sql.DB {
	_ = godotenv.Load()

	host := os.Getenv("DB_HOST")
	if host == "" {
		log.Fatal("DB_HOST is required")
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		log.Fatal("DB_PORT is required")
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		log.Fatal("DB_USER is required")
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		log.Fatal("DB_PASSWORD is required")
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatal("DB_NAME is required")
	}
	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("DB_PORT is invalid: %v", err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbName, sslMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("DB接続設定エラー: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("DB疎通確認エラー (DBが起動していない可能性があります): %v", err)
	}

	fmt.Println("📦 データベースに接続しました。")

	return db
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

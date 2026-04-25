package main

import (
	"database/sql"
	"fmt"
	"log"

	// PostgreSQLのドライバーをインポート (ブランクインポート)
	_ "github.com/lib/pq"
)

// InitDB はデータベースに接続し、必要なテーブルが存在しなければ作成します
func InitDB() *sql.DB {
	// Docker Composeの設定に合わせた接続情報 (DSN)
	// host=db は compose.yaml のサービス名です
	dsn := "host=db port=5432 user=root password=password dbname=budgetMS sslmode=disable"

	// DBへの接続設定
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("DB接続設定エラー: %v", err)
	}

	// 実際にDBと通信できるか確認
	if err := db.Ping(); err != nil {
		log.Fatalf("DB疎通確認エラー (DBが起動していない可能性があります): %v", err)
	}

	fmt.Println("📦 データベースに接続しました。テーブルの初期化を確認します...")

	// ER図を元にしたテーブル作成SQL
	// 外部キー制約(REFERENCES)があるため、作成順序が重要です (親テーブルから先に作る)
	createTablesSQL := `
	-- 1. Userテーブル
	CREATE TABLE IF NOT EXISTS users (
		user_id SERIAL PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		password TEXT NOT NULL,
		login DATE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- 2. Budgetテーブル
	CREATE TABLE IF NOT EXISTS budgets (
		budget_id SERIAL PRIMARY KEY,
		user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
		name TEXT NOT NULL
	);

	-- 3. Categoryテーブル
	CREATE TABLE IF NOT EXISTS categories (
		category_id SERIAL PRIMARY KEY,
		category_name TEXT NOT NULL
	);

	-- 4. Methodテーブル
	CREATE TABLE IF NOT EXISTS methods (
		method_id SERIAL PRIMARY KEY,
		method_name TEXT NOT NULL
	);

	-- 5. Transactionテーブル
	CREATE TABLE IF NOT EXISTS transactions (
		transaction_id SERIAL PRIMARY KEY,
		user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
		date TIMESTAMP NOT NULL,
		amount INT NOT NULL,
		type INT NOT NULL CHECK (type IN (1, 2, 3)), -- 1:income, 2:expense, 3:transfer
		place TEXT,
		note TEXT,
		budget_id INT REFERENCES budgets(budget_id) ON DELETE SET NULL,
		category_id INT REFERENCES categories(category_id) ON DELETE SET NULL,
		method_id INT REFERENCES methods(method_id) ON DELETE SET NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// SQLの実行
	_, err = db.Exec(createTablesSQL)
	if err != nil {
		log.Fatalf("テーブル初期化エラー: %v", err)
	}

	fmt.Println("✅ テーブルの準備が完了しました！")

	return db
}
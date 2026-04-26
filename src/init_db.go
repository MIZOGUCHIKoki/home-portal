package main

import (
	"database/sql"
	"fmt"
	"log"
)

// InitDB は必要なテーブルが存在しなければ作成します
func InitDB(db *sql.DB) {
	fmt.Println("🛠️ テーブルの初期化を確認します...")

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
}
package setup

import (
	"database/sql"
	"fmt"
	"kakeibo/internal/repository"
	"os"
)

func Run(conn *sql.DB) error {
	fmt.Println("🛠️ テーブルの初期化を確認します...")

	createTablesSQL := `
	-- 0. 既存テーブルを削除
	DROP TABLE IF EXISTS advance CASCADE;
	DROP TABLE IF EXISTS transactions CASCADE;
	DROP TABLE IF EXISTS budgets CASCADE;
	DROP TABLE IF EXISTS categories CASCADE;
	DROP TABLE IF EXISTS methods CASCADE;
	DROP TABLE IF EXISTS users CASCADE;

	-- 1. Userテーブル
	CREATE TABLE IF NOT EXISTS users (
		user_id SERIAL PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		password TEXT NOT NULL,
		is_admin BOOLEAN NOT NULL DEFAULT FALSE,
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
		identifier TEXT NOT NULL UNIQUE,
		category_name TEXT NOT NULL
	);

	-- 4. Methodテーブル
	CREATE TABLE IF NOT EXISTS methods (
		method_id SERIAL PRIMARY KEY,
		identifier TEXT NOT NULL UNIQUE,
		method_name TEXT NOT NULL
	);

	-- 5. Transactionテーブル
	CREATE TABLE IF NOT EXISTS transactions (
		transaction_id SERIAL PRIMARY KEY,
		user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
		date TIMESTAMP NOT NULL,
		amount INT NOT NULL, -- 立替を含む金額
		net_amount INT NOT NULL, -- 実質の金額（立替を除いた金額）
		type BOOLEAN NOT NULL, -- true: income, false: expense
		is_transfer BOOLEAN NOT NULL DEFAULT FALSE, -- true: transfer, false: not transfer
		place TEXT,
		note TEXT,
		budget_id INT REFERENCES budgets(budget_id) ON DELETE SET NULL,
		category_id INT REFERENCES categories(category_id) ON DELETE SET NULL,
		method_id INT NOT NULL REFERENCES methods(method_id),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- 6. Advanceテーブル
	CREATE TABLE IF NOT EXISTS advance (
		advance_id SERIAL PRIMARY KEY,
		transaction_id INT REFERENCES transactions(transaction_id) ON DELETE CASCADE,
		amount INT NOT NULL, -- 立替金額
		returned_amount INT NOT NULL DEFAULT 0, -- 返済済みの金額
		name TEXT NOT NULL, -- 立替先の名前
		status BOOLEAN NOT NULL DEFAULT FALSE, -- true: 返済完了, false: 返済未完了
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := conn.Exec(createTablesSQL)
	if err != nil {
		return fmt.Errorf("テーブル初期化エラー: %w", err)
	}

	fmt.Println("✅ テーブルの準備が完了しました！")

	fmt.Println("🌱 マスターデータ（カテゴリ・決済手段・管理者ユーザ）のセットアップを開始します...")

	if _, err := conn.Exec(createTablesSQL); err != nil {
		return fmt.Errorf("テーブル初期化エラー: %w", err)
	}

	if err := repository.SeedDefaultCategories(conn); err != nil {
		return fmt.Errorf("カテゴリ: %w", err)
	}

	if err := repository.SeedDefaultMethods(conn); err != nil {
		return fmt.Errorf("決済手段: %w", err)
	}

	systemadminName := os.Getenv("SYSTEM_USER_ADMIN")
	if systemadminName == "" {
		return fmt.Errorf("SYSTEM_USER_ADMIN is required")
	}
	systemadminEmail := os.Getenv("SYSTEM_USER_ADMIN_EMAIL")
	if systemadminEmail == "" {
		return fmt.Errorf("SYSTEM_USER_ADMIN_EMAIL is required")
	}
	systemadminPassword := os.Getenv("SYSTEM_USER_ADMIN_PASSWORD")
	if systemadminPassword == "" {
		return fmt.Errorf("SYSTEM_USER_ADMIN_PASSWORD is required")
	}

	if err := repository.SeedSystemAdminUser(conn, systemadminName, systemadminEmail, systemadminPassword); err != nil {
		return fmt.Errorf("管理者ユーザ: %w", err)
	}

	fmt.Println("✅ マスターデータのセットアップが完了しました！")
	return nil
}

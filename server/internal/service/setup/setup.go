package setup

import (
	"database/sql"
	"fmt"
	"kakeibo/internal/repository"
	"os"
)

func Run(conn *sql.DB) error {
	fmt.Println("🛠️ テーブルの初期化を確認します...")
	dropTablesSQL := `
	-- 0. 既存テーブルを削除
	DROP TABLE IF EXISTS advance CASCADE;
	DROP TABLE IF EXISTS place_rules CASCADE;
	DROP TABLE IF EXISTS transactions CASCADE;
	DROP TABLE IF EXISTS budgets CASCADE;
	DROP TABLE IF EXISTS categories CASCADE;
	DROP TABLE IF EXISTS methods CASCADE;
	DROP TABLE IF EXISTS users CASCADE;
	`

	if os.Getenv("RESET_DATABASE") == "1" || os.Getenv("RESET_DATABASE") == "true" {
		fmt.Println("⚠️ RESET_DATABASEが有効になっています。既存のテーブルを削除して再作成します...")
		if _, err := conn.Exec(dropTablesSQL); err != nil {
			return fmt.Errorf("テーブル削除エラー: %w", err)
		}
	} else {
		fmt.Println("⚠️ 注意: テーブルの初期化は行われません。RESET_DATABASE環境変数を1またはtrueに設定すると、既存のテーブルが削除されて再作成されます。")
		return nil
	}

	fmt.Println("✅ 既存テーブルの削除が完了しました！")

	createTablesSQL := `
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
		rule BOOLEAN NOT NULL DEFAULT FALSE, -- false: 手動, true: PlaceRuleによる自動設定
		is_csv BOOLEAN NOT NULL DEFAULT FALSE, -- false: 手作業で入力・編集, true: CSVインポート
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

	-- 7. PlaceRuleテーブル（placeの一致判定でidentifier / is_transferを自動判定する）
	-- match_type: 0=含まれる, 1=全文一致, 2=前方一致, 3=後方一致
	CREATE TABLE IF NOT EXISTS place_rules (
		place_rule_id SERIAL PRIMARY KEY,
		text TEXT NOT NULL,
		match_type SMALLINT NOT NULL DEFAULT 0,
		identifier TEXT REFERENCES categories(identifier) ON DELETE SET NULL,
		is_transfer BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE (text, match_type)
	);
	`

	_, err := conn.Exec(createTablesSQL)
	if err != nil {
		return fmt.Errorf("テーブル作成エラー: %w", err)
	}

	fmt.Println("✅ テーブルの準備が完了しました！")

	fmt.Println("🌱 マスターデータ（カテゴリ・決済手段・管理者ユーザ）のセットアップを開始します...")
	if err := repository.SeedDefaultCategories(conn); err != nil {
		return fmt.Errorf("カテゴリ: %w", err)
	}

	if err := repository.SeedDefaultMethods(conn); err != nil {
		return fmt.Errorf("決済手段: %w", err)
	}

	if err := repository.SeedDefaultPlaceRules(conn); err != nil {
		return fmt.Errorf("PlaceRule: %w", err)
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

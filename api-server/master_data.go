package main

import (
	"database/sql"
	"fmt"
	"log"
)

func SeedMasterData(db *sql.DB) {
	fmt.Println("🌱 マスターデータ（カテゴリ・決済手段・管理者ユーザ）のセットアップを開始します...")

	if err := SeedDefaultCategories(db); err != nil {
		log.Fatalf("❌ カテゴリ適用エラー: %v", err)
	}
	fmt.Println("✅ カテゴリ初期データを適用しました")

	if err := SeedDefaultMethods(db); err != nil {
		log.Fatalf("❌ 支払い方法適用エラー: %v", err)
	}
	fmt.Println("✅ 支払い方法初期データを適用しました")

	if err := SeedSystemAdminUser(db); err != nil {
		log.Fatalf("❌ 管理者ユーザ適用エラー: %v", err)
	}
	fmt.Println("✅ 管理者ユーザ初期データを適用しました")

	fmt.Println("✨ マスターデータのセットアップが完了しました！")
}

func SeedSystemAdminUser(db *sql.DB) error {
	systemAdminName := getEnv("SYSTEM_USER_ADMIN", "systemuser")
	systemAdminEmail := getEnv("SYSTEM_USER_ADMIN_EMAIL", systemAdminName+"@local")
	systemAdminPassword := getEnv("SYSTEM_USER_ADMIN_PASSWORD", "admin")

	hash, err := HashPassword(systemAdminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash system admin password: %w", err)
	}

	_, err = db.Exec(
		`INSERT INTO users (email, name, password, is_admin)
		 VALUES ($1, $2, $3, TRUE)
		 ON CONFLICT (email) DO UPDATE
		 SET name = EXCLUDED.name,
		     password = EXCLUDED.password,
		     is_admin = EXCLUDED.is_admin,
		     updated_at = CURRENT_TIMESTAMP`,
		systemAdminEmail,
		systemAdminName,
		hash,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert system admin user: %w", err)
	}

	return nil
}
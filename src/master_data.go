package main

import (
	"database/sql"
	"fmt"
	"log"
)

func SeedMasterData(db *sql.DB) {
	fmt.Println("🌱 マスターデータ（カテゴリ・決済手段）のセットアップを開始します...")

	if err := SeedDefaultCategories(db); err != nil {
		log.Fatalf("❌ カテゴリ適用エラー: %v", err)
	}
	fmt.Println("  ✅ カテゴリ初期データを適用しました")

	if err := SeedDefaultMethods(db); err != nil {
		log.Fatalf("❌ 支払い方法適用エラー: %v", err)
	}
	fmt.Println("  ✅ 支払い方法初期データを適用しました")

	fmt.Println("✨ マスターデータのセットアップが完了しました！")
}
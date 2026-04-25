package main

import (
	"database/sql"
	"fmt"
	"log"
)

// SeedMasterData はカテゴリや支払い方法の初期データをDBに登録します
func SeedMasterData(db *sql.DB) {
	// 1. カテゴリの初期データ（Numbersの表記に合わせています）
	categories := []string{
		"Groceries & Essentials",
		"Dining Out",
		"Convenience Store",
		"Travel & Transit",
		"Subscription & Bills",
		"Shopping & Special",
		"Wages",
		"Allowance",
		"Advance Payment",
		"Out of Category",
	}

	// 2. 支払い方法（口座・クレジットカード・現金）の初期データ
	methods := []string{
		"Cash",
		"JAL Card (JCB)",
		"View Card",
		"みずほ銀行",
		"住信SBI",
	}

	fmt.Println("🌱 マスターデータ（カテゴリ・決済手段）のセットアップを開始します...")

	// カテゴリの挿入
	// （※何度実行してもデータがダブらないように、すでに存在するかチェックしてから入れます）
	for _, cat := range categories {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM categories WHERE category_name = $1)", cat).Scan(&exists)
		if err != nil {
			log.Fatalf("❌ カテゴリの存在チェックエラー: %v", err)
		}

		if !exists {
			_, err = db.Exec("INSERT INTO categories (category_name) VALUES ($1)", cat)
			if err != nil {
				log.Fatalf("❌ カテゴリの挿入エラー (%s): %v", cat, err)
			}
			fmt.Printf("  ✅ カテゴリ追加: %s\n", cat)
		}
	}

	// 支払い方法の挿入
	for _, method := range methods {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM methods WHERE method_name = $1)", method).Scan(&exists)
		if err != nil {
			log.Fatalf("❌ 支払い方法の存在チェックエラー: %v", err)
		}

		if !exists {
			_, err = db.Exec("INSERT INTO methods (method_name) VALUES ($1)", method)
			if err != nil {
				log.Fatalf("❌ 支払い方法の挿入エラー (%s): %v", method, err)
			}
			fmt.Printf("  ✅ 支払い方法追加: %s\n", method)
		}
	}

	fmt.Println("✨ マスターデータのセットアップが完了しました！")
}
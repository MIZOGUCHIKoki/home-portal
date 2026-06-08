package main

import (
	"database/sql"
	"errors"
	"fmt"
)

var DefaultCategories = []string{
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

// Category は categories テーブルのレコードを表す
type Category struct {
	CategoryID   int64
	CategoryName string
}

func CreateCategory(db *sql.DB, name string) (int64, error) {
	if name == "" {
		return 0, errors.New("category name is required")
	}

	var categoryID int64
	err := db.QueryRow(
		"INSERT INTO categories (category_name) VALUES ($1) RETURNING category_id",
		name,
	).Scan(&categoryID)
	if err != nil {
		return 0, err
	}

	return categoryID, nil
}

func GetCategoryByID(db *sql.DB, categoryID int64) (*Category, error) {
	if categoryID <= 0 {
		return nil, errors.New("categoryID is required")
	}

	category := &Category{}
	err := db.QueryRow(
		"SELECT category_id, category_name FROM categories WHERE category_id = $1",
		categoryID,
	).Scan(&category.CategoryID, &category.CategoryName)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func GetCategoryByName(db *sql.DB, name string) (*Category, error) {
	if name == "" {
		return nil, errors.New("category name is required")
	}

	category := &Category{}
	err := db.QueryRow(
		"SELECT category_id, category_name FROM categories WHERE category_name = $1",
		name,
	).Scan(&category.CategoryID, &category.CategoryName)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func ListCategories(db *sql.DB) ([]Category, error) {
	rows, err := db.Query("SELECT category_id, category_name FROM categories ORDER BY category_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Category, 0)
	for rows.Next() {
		var item Category
		if err := rows.Scan(&item.CategoryID, &item.CategoryName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func UpdateCategory(db *sql.DB, categoryID int64, name string) error {
	if categoryID <= 0 {
		return errors.New("categoryID is required")
	}
	if name == "" {
		return errors.New("category name is required")
	}

	result, err := db.Exec(
		"UPDATE categories SET category_name = $1 WHERE category_id = $2",
		name,
		categoryID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func DeleteCategory(db *sql.DB, categoryID int64) error {
	if categoryID <= 0 {
		return errors.New("categoryID is required")
	}

	result, err := db.Exec("DELETE FROM categories WHERE category_id = $1", categoryID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// EnsureCategory は同名カテゴリがなければ作成し，ID を返す
func EnsureCategory(db *sql.DB, name string) (int64, error) {
	item, err := GetCategoryByName(db, name)
	if err == nil {
		return item.CategoryID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	return CreateCategory(db, name)
}

// SeedDefaultCategories はカテゴリの初期データを適用します
func SeedDefaultCategories(db *sql.DB) error {
	for _, name := range DefaultCategories {
		if _, err := EnsureCategory(db, name); err != nil {
			return fmt.Errorf("category apply failed (%s): %w", name, err)
		}
	}

	return nil
}

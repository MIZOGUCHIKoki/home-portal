package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"kakeibo/internal/model"
)

// CreateCategory inserts a new category
func CreateCategory(db *sql.DB, identifier, name string) (int64, error) {
	if identifier == "" {
		return 0, errors.New("identifier is required")
	}
	if name == "" {
		return 0, errors.New("category name is required")
	}

	var categoryID int64

	err := db.QueryRow(
		`INSERT INTO categories (identifier, category_name)
         VALUES ($1, $2)
         RETURNING category_id`,
		identifier,
		name,
	).Scan(&categoryID)

	if err != nil {
		return 0, err
	}

	return categoryID, nil
}

// GetCategoryByID retrieves category by ID
func GetCategoryByID(db *sql.DB, categoryID int64) (*model.Category, error) {
	if categoryID <= 0 {
		return nil, errors.New("categoryID is required")
	}

	category := &model.Category{}

	err := db.QueryRow(
		`SELECT category_id, identifier, category_name
         FROM categories
         WHERE category_id = $1`,
		categoryID,
	).Scan(&category.CategoryID, &category.Identifier, &category.CategoryName)

	if err != nil {
		return nil, err
	}

	return category, nil
}

// GetCategoryByIdentifier retrieves category by identifier
func GetCategoryByIdentifier(db *sql.DB, identifier string) (*model.Category, error) {
	if identifier == "" {
		return nil, errors.New("identifier is required")
	}

	category := &model.Category{}

	err := db.QueryRow(
		`SELECT category_id, identifier, category_name
         FROM categories
         WHERE identifier = $1`,
		identifier,
	).Scan(&category.CategoryID, &category.Identifier, &category.CategoryName)

	if err != nil {
		return nil, err
	}

	return category, nil
}

// ListCategories returns all categories
func ListCategories(db *sql.DB) ([]model.Category, error) {
	rows, err := db.Query(
		`SELECT category_id, identifier, category_name
         FROM categories
         ORDER BY category_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Category

	for rows.Next() {
		var item model.Category

		if err := rows.Scan(
			&item.CategoryID,
			&item.Identifier,
			&item.CategoryName,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

// UpdateCategory updates category name by ID
func UpdateCategory(db *sql.DB, categoryID int64, name string) error {
	if categoryID <= 0 {
		return errors.New("categoryID is required")
	}
	if name == "" {
		return errors.New("category name is required")
	}

	result, err := db.Exec(
		`UPDATE categories
         SET category_name = $1
         WHERE category_id = $2`,
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

// DeleteCategory deletes category by ID
func DeleteCategory(db *sql.DB, categoryID int64) error {
	if categoryID <= 0 {
		return errors.New("categoryID is required")
	}

	result, err := db.Exec(
		`DELETE FROM categories WHERE category_id = $1`,
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

// EnsureCategory ensures category exists by identifier
func EnsureCategory(db *sql.DB, seed model.CategorySeed) (int64, error) {

	item, err := GetCategoryByIdentifier(db, seed.Identifier)
	if err == nil {
		return item.CategoryID, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	return CreateCategory(db, seed.Identifier, seed.CategoryName)
}

// SeedDefaultCategories applies default categories
func SeedDefaultCategories(db *sql.DB) error {
	for _, c := range model.DefaultCategories {
		if _, err := EnsureCategory(db, c); err != nil {
			return fmt.Errorf("category apply failed (%s): %w", c.Identifier, err)
		}
	}
	return nil
}

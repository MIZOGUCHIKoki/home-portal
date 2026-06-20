package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"kakeibo/internal/model"
)

// CreateMethod inserts a new payment method
func CreateMethod(db *sql.DB, identifier, name string) (int64, error) {
    if identifier == "" {
        return 0, errors.New("identifier is required")
    }
    if name == "" {
        return 0, errors.New("method name is required")
    }

    var methodID int64

    err := db.QueryRow(
        `INSERT INTO methods (identifier, method_name)
         VALUES ($1, $2)
         RETURNING method_id`,
        identifier,
        name,
    ).Scan(&methodID)

    if err != nil {
        return 0, err
    }

    return methodID, nil
}

// GetMethodByID retrieves a method by ID
func GetMethodByID(db *sql.DB, methodID int64) (*model.Method, error) {
    if methodID <= 0 {
        return nil, errors.New("methodID is required")
    }

    item := &model.Method{}

    err := db.QueryRow(
        `SELECT method_id, identifier, method_name
         FROM methods
         WHERE method_id = $1`,
        methodID,
    ).Scan(&item.MethodID, &item.Identifier, &item.MethodName)

    if err != nil {
        return nil, err
    }

    return item, nil
}

// GetMethodByIdentifier retrieves a method by identifier (business key)
func GetMethodByIdentifier(db *sql.DB, identifier string) (*model.Method, error) {
    if identifier == "" {
        return nil, errors.New("identifier is required")
    }

    item := &model.Method{}

    err := db.QueryRow(
        `SELECT method_id, identifier, method_name
         FROM methods
         WHERE identifier = $1`,
        identifier,
    ).Scan(&item.MethodID, &item.Identifier, &item.MethodName)

    if err != nil {
        return nil, err
    }

    return item, nil
}

// ListMethods returns all methods
func ListMethods(db *sql.DB) ([]model.Method, error) {
    rows, err := db.Query(
        `SELECT method_id, identifier, method_name
         FROM methods
         ORDER BY method_id`,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var items []model.Method

    for rows.Next() {
        var item model.Method

        if err := rows.Scan(
            &item.MethodID,
            &item.Identifier,
            &item.MethodName,
        ); err != nil {
            return nil, err
        }

        items = append(items, item)
    }

    return items, rows.Err()
}

// UpdateMethod updates method name by ID
func UpdateMethod(db *sql.DB, methodID int64, name string) error {
    if methodID <= 0 {
        return errors.New("methodID is required")
    }
    if name == "" {
        return errors.New("method name is required")
    }

    result, err := db.Exec(
        `UPDATE methods
         SET method_name = $1
         WHERE method_id = $2`,
        name,
        methodID,
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

// DeleteMethod deletes a method by ID
func DeleteMethod(db *sql.DB, methodID int64) error {
    if methodID <= 0 {
        return errors.New("methodID is required")
    }

    result, err := db.Exec(
        `DELETE FROM methods WHERE method_id = $1`,
        methodID,
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

// EnsureMethod ensures a method exists (by identifier)
func EnsureMethod(db *sql.DB, seed model.MethodSeed) (int64, error) {

    item, err := GetMethodByIdentifier(db, seed.Identifier)
    if err == nil {
        return item.MethodID, nil
    }

    if !errors.Is(err, sql.ErrNoRows) {
        return 0, err
    }

    return CreateMethod(db, seed.Identifier, seed.MethodName)
}

// SeedDefaultMethods applies initial master data
func SeedDefaultMethods(db *sql.DB) error {
    for _, m := range model.DefaultMethods {
        if _, err := EnsureMethod(db, m); err != nil {
            return fmt.Errorf("method apply failed (%s): %w", m.Identifier, err)
        }
    }
    return nil
}
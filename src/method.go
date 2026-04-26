package main

import (
	"database/sql"
	"errors"
	"fmt"
)

var DefaultMethods = []string{
	"Cash",
	"JAL Card (JCB)",
	"JAL Card (Master)",
	"JAL Card (AMEX)",
	"View Card",
	"みずほ銀行",
	"住信SBI",
}

// PaymentMethod は methods テーブルのレコードを表します
type PaymentMethod struct {
	MethodID   int64
	MethodName string
}

func CreateMethod(db *sql.DB, name string) (int64, error) {
	if name == "" {
		return 0, errors.New("method name is required")
	}

	var methodID int64
	err := db.QueryRow(
		"INSERT INTO methods (method_name) VALUES ($1) RETURNING method_id",
		name,
	).Scan(&methodID)
	if err != nil {
		return 0, err
	}

	return methodID, nil
}

func GetMethodByID(db *sql.DB, methodID int64) (*PaymentMethod, error) {
	if methodID <= 0 {
		return nil, errors.New("methodID is required")
	}

	item := &PaymentMethod{}
	err := db.QueryRow(
		"SELECT method_id, method_name FROM methods WHERE method_id = $1",
		methodID,
	).Scan(&item.MethodID, &item.MethodName)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func GetMethodByName(db *sql.DB, name string) (*PaymentMethod, error) {
	if name == "" {
		return nil, errors.New("method name is required")
	}

	item := &PaymentMethod{}
	err := db.QueryRow(
		"SELECT method_id, method_name FROM methods WHERE method_name = $1",
		name,
	).Scan(&item.MethodID, &item.MethodName)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func ListMethods(db *sql.DB) ([]PaymentMethod, error) {
	rows, err := db.Query("SELECT method_id, method_name FROM methods ORDER BY method_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]PaymentMethod, 0)
	for rows.Next() {
		var item PaymentMethod
		if err := rows.Scan(&item.MethodID, &item.MethodName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func UpdateMethod(db *sql.DB, methodID int64, name string) error {
	if methodID <= 0 {
		return errors.New("methodID is required")
	}
	if name == "" {
		return errors.New("method name is required")
	}

	result, err := db.Exec(
		"UPDATE methods SET method_name = $1 WHERE method_id = $2",
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

func DeleteMethod(db *sql.DB, methodID int64) error {
	if methodID <= 0 {
		return errors.New("methodID is required")
	}

	result, err := db.Exec("DELETE FROM methods WHERE method_id = $1", methodID)
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

// EnsureMethod は同名の支払い方法がなければ作成し、ID を返します
func EnsureMethod(db *sql.DB, name string) (int64, error) {
	item, err := GetMethodByName(db, name)
	if err == nil {
		return item.MethodID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	return CreateMethod(db, name)
}

// SeedDefaultMethods は支払い方法の初期データを適用します
func SeedDefaultMethods(db *sql.DB) error {
	for _, name := range DefaultMethods {
		if _, err := EnsureMethod(db, name); err != nil {
			return fmt.Errorf("method apply failed (%s): %w", name, err)
		}
	}

	return nil
}

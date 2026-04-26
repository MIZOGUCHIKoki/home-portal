package main

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// users テーブルのレコード
type User struct {
	UserID    int64
	Email     string
	Name      string
	Password  string
	IsAdmin   bool
	Login     sql.NullTime
	CreatedAt time.Time
	UpdatedAt time.Time
}

func CreateUser(db *sql.DB, email, name, password string) (int64, error) {
	return CreateUserWithAdmin(db, email, name, password, false)
}

// CreateUserWithAdmin は管理者フラグ付きでユーザを作成します
func CreateUserWithAdmin(db *sql.DB, email, name, password string, isAdmin bool) (int64, error) {
	if email == "" {
		return 0, errors.New("email is required")
	}
	if name == "" {
		return 0, errors.New("name is required")
	}
	if password == "" {
		return 0, errors.New("password is required")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return 0, err
	}

	var userID int64
	err = db.QueryRow(
		"INSERT INTO users (email, name, password, is_admin) VALUES ($1, $2, $3, $4) RETURNING user_id",
		email,
		name,
		hash,
		isAdmin,
	).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// user_id でユーザを取得します
func GetUserByID(db *sql.DB, userID int64) (*User, error) {
	user := &User{}
	err := db.QueryRow(
		"SELECT user_id, email, name, password, is_admin, login, created_at, updated_at FROM users WHERE user_id = $1",
		userID,
	).Scan(
		&user.UserID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.IsAdmin,
		&user.Login,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// email でユーザを取得
func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	user := &User{}
	err := db.QueryRow(
		"SELECT user_id, email, name, password, is_admin, login, created_at, updated_at FROM users WHERE email = $1",
		email,
	).Scan(
		&user.UserID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.IsAdmin,
		&user.Login,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// 全ユーザを取得
func ListUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query(
		"SELECT user_id, email, name, password, is_admin, login, created_at, updated_at FROM users ORDER BY user_id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(
			&user.UserID,
			&user.Email,
			&user.Name,
			&user.Password,
			&user.IsAdmin,
			&user.Login,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// 既存ユーザの情報を更新
func UpdateUser(db *sql.DB, userID int64, email, name, password string) error {
	if userID <= 0 {
		return errors.New("userID is required")
	}
	if email == "" {
		return errors.New("email is required")
	}
	if name == "" {
		return errors.New("name is required")
	}
	if password == "" {
		return errors.New("password is required")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	result, err := db.Exec(
		"UPDATE users SET email = $1, name = $2, password = $3, updated_at = CURRENT_TIMESTAMP WHERE user_id = $4",
		email,
		name,
		hash,
		userID,
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

// UpdateUserWithAdmin は管理者フラグも含めてユーザを更新します
func UpdateUserWithAdmin(db *sql.DB, userID int64, email, name, password string, isAdmin bool) error {
	if userID <= 0 {
		return errors.New("userID is required")
	}
	if email == "" {
		return errors.New("email is required")
	}
	if name == "" {
		return errors.New("name is required")
	}
	if password == "" {
		return errors.New("password is required")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	result, err := db.Exec(
		"UPDATE users SET email = $1, name = $2, password = $3, is_admin = $4, updated_at = CURRENT_TIMESTAMP WHERE user_id = $5",
		email,
		name,
		hash,
		isAdmin,
		userID,
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

// 平文パスワードを bcrypt でハッシュ化
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// 平文パスワードとハッシュを照合
func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// DeleteUser はユーザを削除
func DeleteUser(db *sql.DB, userID int64) error {
	if userID <= 0 {
		return errors.New("userID is required")
	}

	result, err := db.Exec("DELETE FROM users WHERE user_id = $1", userID)
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

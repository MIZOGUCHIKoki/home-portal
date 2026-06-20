package repository

import (
    "database/sql"
    "errors"
    "fmt"
    "kakeibo/internal/model"
    "kakeibo/internal/service/auth"
)

func CreateUser(db *sql.DB, email, name, hashedPassword string, isAdmin bool) (int64, error) {
    if email == "" {
        return 0, errors.New("email is required")
    }
    if name == "" {
        return 0, errors.New("name is required")
    }
    if hashedPassword == "" {
        return 0, errors.New("password is required")
    }

    var userID int64

    err := db.QueryRow(
        "INSERT INTO users (email, name, password, is_admin) VALUES ($1, $2, $3, $4) RETURNING user_id",
        email,
        name,
        hashedPassword,
        isAdmin,
    ).Scan(&userID)

    if err != nil {
        return 0, err
    }

    return userID, nil
}

func GetUserByID(db *sql.DB, userID int64) (*model.User, error) {
    user := &model.User{}

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

func GetUserByEmail(db *sql.DB, email string) (*model.User, error) {
    user := &model.User{}

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

func ListUsers(db *sql.DB) ([]model.User, error) {
    rows, err := db.Query(
        "SELECT user_id, email, name, password, is_admin, login, created_at, updated_at FROM users ORDER BY user_id",
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []model.User

    for rows.Next() {
        var user model.User
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

    return users, rows.Err()
}

func UpdateUser(db *sql.DB, userID int64, email, name, hashedPassword string, isAdmin bool) error {
    result, err := db.Exec(
        "UPDATE users SET email=$1, name=$2, password=$3, is_admin=$4, updated_at=CURRENT_TIMESTAMP WHERE user_id=$5",
        email,
        name,
        hashedPassword,
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

func DeleteUser(db *sql.DB, userID int64) error {
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

func SeedSystemAdminUser(db *sql.DB, systemAdminName, systemAdminEmail, systemAdminPassword string) error {
	hash, err := auth.HashPassword(systemAdminPassword)
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
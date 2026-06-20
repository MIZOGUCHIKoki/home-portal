package model

import (
	"database/sql"
	"time"
)

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

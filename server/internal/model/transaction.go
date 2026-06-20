package model

import "time"

type Transaction struct {
	TransactionID int64
	UserID        int64
	Date          time.Time

	Amount    int
	NetAmount int

	Type       bool
	IsTransfer bool

	// method は NOT NULL
	MethodID int64

	// nullable
	CategoryID *int64
	Place      *string
	Note       *string
}

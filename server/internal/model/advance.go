package model

import "time"

type Advance struct {
	AdvanceID     int64
	TransactionID int64

	Name           string
	Amount         int
	ReturnedAmount int
	Status         bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

package dto

type UpdateAdvanceRequest struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Amount int    `json:"amount"`
}

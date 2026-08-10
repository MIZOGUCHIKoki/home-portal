package dto

type AdvanceInput struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"`
}

type CreateTransactionRequest struct {
	UserID int64  `json:"user_id"`
	Date   string `json:"date"`

	Amount    int `json:"amount"`
	NetAmount int `json:"net_amount"`

	Type       bool `json:"type"`
	IsTransfer bool `json:"is_transfer"`

	Place string `json:"place"`
	Note  string `json:"note"`

	MethodID   int64  `json:"method_id"`
	CategoryID *int64 `json:"category_id"`

	Advances []AdvanceInput `json:"advances"`

	RefundAdvanceID *int64 `json:"refund_advance_id"`
}

type UpdateTransactionRequest struct {
	ID   int64  `json:"id"`
	Date string `json:"date"`

	Amount    int `json:"amount"`
	NetAmount int `json:"net_amount"`

	Type       bool `json:"type"`
	IsTransfer bool `json:"is_transfer"`

	Place string `json:"place"`
	Note  string `json:"note"`

	MethodID   int64  `json:"method_id"`
	CategoryID *int64 `json:"category_id"`
}

type AdvanceResponse struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Amount         int    `json:"amount"`
	ReturnedAmount int    `json:"returned_amount"`
	Status         bool   `json:"status"`
}

type TransactionResponse struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	Date       string `json:"date"`
	Amount     int    `json:"amount"`
	NetAmount  int    `json:"net_amount"`
	Type       bool   `json:"type"`
	IsTransfer bool   `json:"is_transfer"`

	MethodID   int64  `json:"method_id"`
	CategoryID *int64 `json:"category_id"`

	Place string `json:"place"`
	Note  string `json:"note"`

	// Rule: false=手動, true=PlaceRuleによる自動設定
	Rule bool `json:"rule"`

	// IsCsv: 0=手作業(未確認), 1=CSVインポート, 2=手作業(CSV突合済み)
	IsCsv int `json:"is_csv"`

	Advances []AdvanceResponse `json:"advances"`
}

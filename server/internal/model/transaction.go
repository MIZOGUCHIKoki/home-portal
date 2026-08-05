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

	// Rule: category/is_transferの設定方法 (false: 手動, true: PlaceRuleによる自動設定)
	Rule bool

	// IsCsv: 取引の登録経路 (false: 手作業で入力・編集, true: CSVインポート)
	IsCsv bool
}

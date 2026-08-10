package model

import "time"

// CsvStatus decides how a transaction was registered and, for manual
// entries, whether it has since been reconciled against a CSV import.
type CsvStatus int

const (
	CsvStatusManual     CsvStatus = 0 // 手作業で入力・編集、CSVとの突合未確認
	CsvStatusImported   CsvStatus = 1 // CSVインポートによる新規登録
	CsvStatusReconciled CsvStatus = 2 // 手入力だが、CSVインポート時に一致確認済み
)

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

	// IsCsv: 取引の登録経路・突合状態 (0: 手作業/未確認, 1: CSVインポート, 2: 手作業/CSV突合済み)
	IsCsv CsvStatus
}

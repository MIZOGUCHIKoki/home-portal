package model

// MonthlySummary aggregates income/expense totals for one calendar month
// (transfers are excluded).
type MonthlySummary struct {
	YearMonth string // "2026-01"
	Income    int
	Expense   int
}

// CategorySummary aggregates income/expense totals for one category
// (transfers are excluded).
type CategorySummary struct {
	CategoryID   int64
	Identifier   string
	CategoryName string
	Income       int
	Expense      int
	Count        int
}

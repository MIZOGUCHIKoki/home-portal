package dto

type MonthlySummaryResponse struct {
	YearMonth string `json:"year_month"`
	Income    int    `json:"income"`
	Expense   int    `json:"expense"`
	Net       int    `json:"net"`
}

type CategorySummaryResponse struct {
	CategoryID   int64  `json:"category_id"`
	Identifier   string `json:"identifier"`
	CategoryName string `json:"category_name"`
	Income       int    `json:"income"`
	Expense      int    `json:"expense"`
	Count        int    `json:"count"`
}

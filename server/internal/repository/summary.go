package repository

import (
	"database/sql"
	"strconv"

	"kakeibo/internal/model"
)

// GetMonthlyTransactionSummary returns income/expense totals per month
// (transfers excluded). If year is non-nil, only that year is returned.
func GetMonthlyTransactionSummary(db *sql.DB, userID int64, year *int) ([]model.MonthlySummary, error) {
	query := `
    SELECT
        to_char(date, 'YYYY-MM') AS year_month,
        COALESCE(SUM(CASE WHEN type THEN net_amount ELSE 0 END), 0) AS income,
        COALESCE(SUM(CASE WHEN type = false THEN net_amount ELSE 0 END), 0) AS expense
    FROM transactions
    WHERE user_id = $1 AND is_transfer = false
    `
	args := []any{userID}

	if year != nil {
		query += " AND EXTRACT(YEAR FROM date) = $2"
		args = append(args, *year)
	}

	query += `
    GROUP BY year_month
    ORDER BY year_month
    `

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.MonthlySummary

	for rows.Next() {
		var item model.MonthlySummary
		if err := rows.Scan(&item.YearMonth, &item.Income, &item.Expense); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// GetCategoryTransactionSummary returns income/expense totals per category
// (transfers, and transactions without a category, excluded). year/month
// narrow the period when non-nil; month is only applied when year is also
// set.
func GetCategoryTransactionSummary(db *sql.DB, userID int64, year, month *int) ([]model.CategorySummary, error) {
	query := `
    SELECT
        c.category_id,
        c.identifier,
        c.category_name,
        COALESCE(SUM(CASE WHEN t.type THEN t.net_amount ELSE 0 END), 0) AS income,
        COALESCE(SUM(CASE WHEN t.type = false THEN t.net_amount ELSE 0 END), 0) AS expense,
        COUNT(t.transaction_id) AS count
    FROM transactions t
    JOIN categories c ON c.category_id = t.category_id
    WHERE t.user_id = $1 AND t.is_transfer = false
    `
	args := []any{userID}

	if year != nil {
		args = append(args, *year)
		query += " AND EXTRACT(YEAR FROM t.date) = $" + strconv.Itoa(len(args))

		if month != nil {
			args = append(args, *month)
			query += " AND EXTRACT(MONTH FROM t.date) = $" + strconv.Itoa(len(args))
		}
	}

	query += `
    GROUP BY c.category_id, c.identifier, c.category_name
    ORDER BY expense DESC, income DESC
    `

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.CategorySummary

	for rows.Next() {
		var item model.CategorySummary
		if err := rows.Scan(&item.CategoryID, &item.Identifier, &item.CategoryName, &item.Income, &item.Expense, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

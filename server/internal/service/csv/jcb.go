package csv

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ImportJCB imports all CSV files in a directory
func ImportJCB(db *sql.DB, dirPath string, userID int) error {
	return importCSVDir(db, dirPath, userID, "JCB", "jal_jcb", parseJCBRow)
}

func parseJCBRow(record []string) (csvRow, bool, error) {
	// ✅ 列数不足
	if len(record) < 5 {
		return csvRow{}, false, nil
	}

	// ✅ ヘッダ行スキップ
	if strings.Contains(record[2], "ご利用日") {
		return csvRow{}, false, nil
	}

	dateStr := strings.TrimSpace(record[2])
	place := strings.TrimSpace(record[3])
	amountStr := strings.ReplaceAll(record[4], ",", "")

	date, err := time.Parse("2006/01/02", dateStr)
	if err != nil {
		return csvRow{}, false, fmt.Errorf("日付エラー: %s: %w", dateStr, err)
	}

	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return csvRow{}, false, fmt.Errorf("金額エラー: %s: %w", amountStr, err)
	}

	row := csvRow{Date: date, Place: place}

	// ✅ 収入 / 支出判定（列の符号で自動判定）
	if amount < 0 {
		row.Type = true
		row.Amount = -amount
	} else {
		row.Type = false
		row.Amount = amount
	}

	return row, true, nil
}

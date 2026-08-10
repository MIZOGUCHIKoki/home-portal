package csv

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ImportSBINetBank imports ALL CSV files in a directory
func ImportSBINetBank(db *sql.DB, dirPath string, userID int) error {
	return importCSVDir(db, dirPath, userID, "SBI銀行", "0038", parseNetBKRow) // ドコモSMTBネット銀行
}

func parseNetBKRow(record []string) (csvRow, bool, error) {
	// ✅ 列数チェック
	if len(record) < 5 {
		return csvRow{}, false, nil
	}

	// ✅ ヘッダスキップ
	if strings.Contains(record[0], "日付") {
		return csvRow{}, false, nil
	}

	dateStr := strings.TrimSpace(record[0])
	content := strings.TrimSpace(record[1])

	debitStr := strings.ReplaceAll(record[2], ",", "")
	creditStr := strings.ReplaceAll(record[3], ",", "")

	date, err := time.Parse("2006/01/02", dateStr)
	if err != nil {
		return csvRow{}, false, fmt.Errorf("日付エラー: %s: %w", dateStr, err)
	}

	row := csvRow{Date: date, Place: content}

	switch {
	case debitStr != "": // ✅ 出金（支出）
		amount, err := strconv.Atoi(debitStr)
		if err != nil {
			return csvRow{}, false, fmt.Errorf("出金パースエラー: %s: %w", debitStr, err)
		}
		row.Type = false
		row.Amount = amount

	case creditStr != "": // ✅ 入金（収入）
		amount, err := strconv.Atoi(creditStr)
		if err != nil {
			return csvRow{}, false, fmt.Errorf("入金パースエラー: %s: %w", creditStr, err)
		}
		row.Type = true
		row.Amount = amount

	default:
		return csvRow{}, false, nil
	}

	return row, true, nil
}
